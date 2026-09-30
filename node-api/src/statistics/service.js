import { ApiError } from '../http/errors.js';

// Una matriz cuenta como diagonal cuando todo lo que queda fuera de la diagonal
// principal es cero. La tolerancia absorbe el redondeo de la factorización QR.
const DIAGONAL_TOLERANCE = 1e-9;

export function normalizeMatrices(body) {
  const source = body?.matrices ?? body?.matrix;

  if (source === undefined || source === null) {
    throw invalid('el cuerpo debe incluir un campo matrices o matrix');
  }

  if (looksLikeMatrix(source)) {
    return new Map([['matrix', assertMatrix(source, 'matrix')]]);
  }

  if (Array.isArray(source)) {
    if (source.length === 0) {
      throw invalid('matrices no puede estar vacío');
    }
    return new Map(source.map((matrix, index) => [`matrix[${index}]`, assertMatrix(matrix, `matrices[${index}]`)]));
  }

  if (typeof source === 'object') {
    const entries = Object.entries(source);
    if (entries.length === 0) {
      throw invalid('matrices no puede estar vacío');
    }
    return new Map(entries.map(([name, matrix]) => [name, assertMatrix(matrix, `matrices.${name}`)]));
  }

  throw invalid('matrices debe ser una matriz, una lista de matrices o un objeto de matrices');
}

export function computeStatistics(matrices) {
  let count = 0;
  let sum = 0;
  let min = Infinity;
  let max = -Infinity;

  const diagonal = { any: false, matrices: {} };

  for (const [name, matrix] of matrices) {
    for (const row of matrix) {
      for (const value of row) {
        count += 1;
        sum += value;
        min = Math.min(min, value);
        max = Math.max(max, value);
      }
    }

    const flag = isDiagonal(matrix);
    diagonal.matrices[name] = flag;
    diagonal.any = diagonal.any || flag;
  }

  return {
    count,
    sum,
    average: count === 0 ? 0 : sum / count,
    min,
    max,
    diagonal,
  };
}

function isDiagonal(matrix) {
  const size = matrix.length;
  if (matrix[0].length !== size) {
    return false;
  }

  for (let row = 0; row < size; row += 1) {
    for (let column = 0; column < size; column += 1) {
      if (row !== column && Math.abs(matrix[row][column]) > DIAGONAL_TOLERANCE) {
        return false;
      }
    }
  }

  return true;
}

function looksLikeMatrix(value) {
  return Array.isArray(value) && value.length > 0 && Array.isArray(value[0]) && typeof value[0][0] === 'number';
}

function assertMatrix(value, label) {
  if (!Array.isArray(value) || value.length === 0) {
    throw invalid(`${label} debe ser un arreglo de filas no vacío`);
  }
  if (!Array.isArray(value[0]) || value[0].length === 0) {
    throw invalid(`${label} debe contener filas no vacías`);
  }

  const columns = value[0].length;
  value.forEach((row, index) => {
    if (!Array.isArray(row)) {
      throw invalid(`${label}[${index}] debe ser un arreglo`);
    }
    if (row.length !== columns) {
      throw invalid(`${label}[${index}] tiene ${row.length} valores, se esperaban ${columns}`);
    }
    row.forEach((entry, column) => {
      if (typeof entry !== 'number' || !Number.isFinite(entry)) {
        throw invalid(`${label}[${index}][${column}] debe ser un número finito`);
      }
    });
  });

  return value;
}

function invalid(message) {
  return new ApiError(400, 'invalid_matrix', message);
}
