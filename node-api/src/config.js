import { parseUsers } from './auth/credentials.js';

export function loadConfig(env = process.env) {
  const jwtSecret = env.JWT_SECRET ?? '';
  if (!jwtSecret) {
    throw new Error('JWT_SECRET es obligatorio');
  }

  return {
    port: readPort(env.PORT, 3000),
    jwtSecret,
    jwtIssuer: env.JWT_ISSUER || 'reto-tecnico',
    jwtExpiresIn: env.JWT_EXPIRES_IN || '1h',
    users: parseUsers(env.AUTH_USERS),
  };
}

function readPort(raw, fallback) {
  if (raw === undefined || raw === '') {
    return fallback;
  }

  const port = Number(raw);
  if (!Number.isInteger(port) || port < 1 || port > 65535) {
    throw new Error(`PORT debe ser un número de puerto válido, se recibió ${raw}`);
  }

  return port;
}
