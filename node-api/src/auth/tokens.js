import jwt from 'jsonwebtoken';

export const ALGORITHM = 'HS256';

export function signToken(config, subject) {
  return jwt.sign({}, config.jwtSecret, {
    subject,
    issuer: config.jwtIssuer,
    algorithm: ALGORITHM,
    expiresIn: config.jwtExpiresIn,
  });
}

export function verifyToken(config, token) {
  return jwt.verify(token, config.jwtSecret, {
    issuer: config.jwtIssuer,
    algorithms: [ALGORITHM],
  });
}
