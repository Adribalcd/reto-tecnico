import { Router } from 'express';

import { ApiError } from '../http/errors.js';
import { authenticate } from './credentials.js';
import { signToken } from './tokens.js';

export function createAuthRouter({ config }) {
  const router = Router();

  router.post('/login', (req, res, next) => {
    const { username, password } = req.body ?? {};

    if (typeof username !== 'string' || typeof password !== 'string') {
      return next(new ApiError(400, 'invalid_body', 'usuario y contraseña deben ser cadenas'));
    }

    if (!authenticate(config.users, username, password)) {
      return next(new ApiError(401, 'invalid_credentials', 'el usuario o la contraseña son incorrectos'));
    }

    return res.json({
      token: signToken(config, username),
      tokenType: 'Bearer',
      expiresIn: config.jwtExpiresIn,
    });
  });

  return router;
}
