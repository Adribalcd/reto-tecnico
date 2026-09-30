const NODE_API = '/node';
const GO_API = '/go';

const EXAMPLE = [
  [12, -51, 4],
  [6, 167, -68],
  [-4, 24, -41],
];

const state = {
  token: null,
  values: [],
};

const elements = {
  loginForm: document.getElementById('login-form'),
  username: document.getElementById('username'),
  password: document.getElementById('password'),
  loginMessage: document.getElementById('login-message'),
  session: document.getElementById('session'),
  rows: document.getElementById('rows'),
  cols: document.getElementById('cols'),
  fillExample: document.getElementById('fill-example'),
  compute: document.getElementById('compute'),
  matrixInput: document.getElementById('matrix-input'),
  matrixMessage: document.getElementById('matrix-message'),
  results: document.getElementById('results'),
  matrixQ: document.getElementById('matrix-q'),
  matrixR: document.getElementById('matrix-r'),
  statistics: document.getElementById('statistics'),
};

elements.loginForm.addEventListener('submit', async (event) => {
  event.preventDefault();
  setMessage(elements.loginMessage, 'Validando credenciales...');

  try {
    const response = await fetch(`${NODE_API}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: elements.username.value,
        password: elements.password.value,
      }),
    });
    const payload = await response.json();

    if (!response.ok) {
      throw new Error(payload?.error?.message ?? 'No se pudo iniciar sesión.');
    }

    state.token = payload.token;
    elements.session.textContent = `Sesión: ${elements.username.value}`;
    elements.session.dataset.state = 'authenticated';
    elements.compute.disabled = false;
    setMessage(elements.loginMessage, 'Sesión iniciada. Las APIs están habilitadas.', 'success');
  } catch (error) {
    state.token = null;
    elements.compute.disabled = true;
    elements.session.textContent = 'Sin sesión';
    elements.session.dataset.state = 'anonymous';
    setMessage(elements.loginMessage, error.message, 'error');
  }
});

elements.rows.addEventListener('change', rebuildMatrix);
elements.cols.addEventListener('change', rebuildMatrix);

elements.fillExample.addEventListener('click', () => {
  elements.rows.value = 3;
  elements.cols.value = 3;
  state.values = EXAMPLE.map((row) => row.map(String));
  buildMatrixInput();
});

elements.compute.addEventListener('click', async () => {
  if (!state.token) {
    setMessage(elements.matrixMessage, 'Inicia sesión antes de calcular.', 'error');
    return;
  }

  const matrix = readMatrix();
  if (matrix instanceof Error) {
    setMessage(elements.matrixMessage, matrix.message, 'error');
    return;
  }

  elements.compute.disabled = true;
  setMessage(elements.matrixMessage, 'Calculando...');

  try {
    const response = await fetch(`${GO_API}/api/v1/qr`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${state.token}`,
      },
      body: JSON.stringify({ matrix }),
    });
    const payload = await response.json();

    if (response.status === 401) {
      state.token = null;
      elements.compute.disabled = true;
      throw new Error('El token expiró. Inicia sesión de nuevo.');
    }
    if (!response.ok) {
      throw new Error(payload?.error?.message ?? 'La factorización falló.');
    }

    renderResults(payload);
    setMessage(elements.matrixMessage, 'Factorización calculada.', 'success');
  } catch (error) {
    setMessage(elements.matrixMessage, error.message, 'error');
  } finally {
    if (state.token) {
      elements.compute.disabled = false;
    }
  }
});

function rebuildMatrix() {
  elements.rows.value = clampDimension(elements.rows.value);
  elements.cols.value = clampDimension(elements.cols.value);
  buildMatrixInput();
}

function clampDimension(raw) {
  const value = Number.parseInt(raw, 10);
  if (Number.isNaN(value)) {
    return 1;
  }
  return Math.min(8, Math.max(1, value));
}

function buildMatrixInput() {
  const rows = clampDimension(elements.rows.value);
  const cols = clampDimension(elements.cols.value);

  elements.matrixInput.replaceChildren();

  for (let row = 0; row < rows; row += 1) {
    const line = document.createElement('div');
    line.className = 'matrix-row';

    for (let col = 0; col < cols; col += 1) {
      const input = document.createElement('input');
      input.type = 'number';
      input.step = 'any';
      input.dataset.row = row;
      input.dataset.col = col;
      input.value = state.values[row]?.[col] ?? '0';
      line.append(input);
    }

    elements.matrixInput.append(line);
  }

  state.values = readCells();
}

function readCells() {
  const rows = clampDimension(elements.rows.value);
  const cols = clampDimension(elements.cols.value);
  const values = [];

  for (let row = 0; row < rows; row += 1) {
    values[row] = [];
    for (let col = 0; col < cols; col += 1) {
      const input = elements.matrixInput.querySelector(`input[data-row="${row}"][data-col="${col}"]`);
      values[row][col] = input?.value ?? '0';
    }
  }

  return values;
}

function readMatrix() {
  const cells = readCells();
  const matrix = [];

  for (let row = 0; row < cells.length; row += 1) {
    matrix[row] = [];
    for (let col = 0; col < cells[row].length; col += 1) {
      const value = Number.parseFloat(cells[row][col]);
      if (Number.isNaN(value)) {
        return new Error(`El valor de la fila ${row + 1}, columna ${col + 1} no es un número.`);
      }
      matrix[row][col] = value;
    }
  }

  state.values = cells;
  return matrix;
}

function renderResults(payload) {
  renderMatrix(elements.matrixQ, payload.q);
  renderMatrix(elements.matrixR, payload.r);
  renderStatistics(elements.statistics, payload.statistics);
  elements.results.hidden = false;
}

function renderMatrix(container, matrix) {
  const table = document.createElement('table');

  for (const row of matrix) {
    const tr = document.createElement('tr');
    for (const value of row) {
      const td = document.createElement('td');
      td.textContent = formatNumber(value);
      tr.append(td);
    }
    table.append(tr);
  }

  container.replaceChildren(table);
}

function renderStatistics(container, statistics) {
  container.replaceChildren();

  const entries = [
    ['Máximo', formatNumber(statistics.max)],
    ['Mínimo', formatNumber(statistics.min)],
    ['Promedio', formatNumber(statistics.average)],
    ['Suma total', formatNumber(statistics.sum)],
    ['Valores', String(statistics.count)],
    ['Alguna matriz diagonal', statistics.diagonal.any ? 'Sí' : 'No'],
  ];

  for (const [label, value] of entries) {
    const wrapper = document.createElement('div');
    const term = document.createElement('dt');
    const description = document.createElement('dd');

    term.textContent = label;
    description.textContent = value;
    wrapper.append(term, description);
    container.append(wrapper);
  }
}

function formatNumber(value) {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    return '—';
  }
  if (value === 0) {
    return '0';
  }
  return String(Number.parseFloat(value.toPrecision(6)));
}

function setMessage(element, text, tone) {
  element.textContent = text;
  if (tone) {
    element.dataset.tone = tone;
  } else {
    delete element.dataset.tone;
  }
}

buildMatrixInput();
