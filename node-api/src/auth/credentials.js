import { timingSafeEqual } from 'node:crypto';

export function parseUsers(raw) {
  const users = new Map();

  for (const entry of String(raw ?? '').split(',')) {
    const value = entry.trim();
    if (!value) {
      continue;
    }

    const separator = value.indexOf(':');
    if (separator <= 0) {
      continue;
    }

    users.set(value.slice(0, separator), value.slice(separator + 1));
  }

  return users;
}

// authenticate compara la contraseña recibida con la almacenada sin filtrar
// longitud ni contenido por tiempos de respuesta.
export function authenticate(users, username, password) {
  const expected = users.get(String(username ?? '')) ?? '';
  if (expected === '') {
    return false;
  }

  return safeEqual(String(password ?? ''), expected);
}

function safeEqual(a, b) {
  const left = Buffer.from(a);
  const right = Buffer.from(b);

  if (left.length !== right.length) {
    timingSafeEqual(left, left);
    return false;
  }

  return timingSafeEqual(left, right);
}
