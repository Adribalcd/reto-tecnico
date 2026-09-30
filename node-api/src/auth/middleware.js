import { ApiError } from '../http/errors.js';
import { verifyToken } from './tokens.js';

export function requireAuth(config) {
  return (req, res, next) => {
    const [scheme, token] = (req.get('authorization') ?? '').split(' ');

    if (scheme !== 'Bearer' || !token) {
      return next(new ApiError(401, 'unauthorized', 'falta el header Authorization con un token Bearer'));
    }

    try {
      req.user = verifyToken(config, token);
      return next();
    } catch {
      return next(new ApiError(401, 'unauthorized', 'el token no es válido o ya expiró'));
    }
  };
}
