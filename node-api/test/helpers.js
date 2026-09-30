import { createApp } from '../src/app.js';

export const testConfig = {
  jwtSecret: 'test-secret',
  jwtIssuer: 'reto-tecnico',
  jwtExpiresIn: '1h',
  users: new Map([
    ['admin', 'admin123'],
    ['analista', 'analista123'],
  ]),
};

export function createTestApp(overrides = {}) {
  return createApp({ config: { ...testConfig, ...overrides } });
}
