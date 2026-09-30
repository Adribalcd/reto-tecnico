import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import request from 'supertest';

import { createTestApp } from './helpers.js';

async function login(app, username = 'admin', password = 'admin123') {
  const response = await request(app).post('/auth/login').send({ username, password });
  return response.body.token;
}

describe('POST /auth/login', () => {
  it('emite un token con credenciales válidas', async () => {
    const response = await request(createTestApp()).post('/auth/login').send({
      username: 'admin',
      password: 'admin123',
    });

    assert.equal(response.status, 200);
    assert.equal(response.body.tokenType, 'Bearer');
    assert.equal(typeof response.body.token, 'string');
    assert.ok(response.body.token.split('.').length === 3);
  });

  it('rechaza credenciales incorrectas', async () => {
    const response = await request(createTestApp()).post('/auth/login').send({
      username: 'admin',
      password: 'incorrecta',
    });

    assert.equal(response.status, 401);
    assert.equal(response.body.error.code, 'invalid_credentials');
  });

  it('rechaza un cuerpo mal formado', async () => {
    const response = await request(createTestApp()).post('/auth/login').send({ username: 'admin' });

    assert.equal(response.status, 400);
    assert.equal(response.body.error.code, 'invalid_body');
  });
});

describe('GET /health', () => {
  it('informa el estado del servicio', async () => {
    const response = await request(createTestApp()).get('/health');

    assert.equal(response.status, 200);
    assert.deepEqual(response.body, { status: 'ok', service: 'node-api' });
  });
});

describe('POST /api/v1/statistics', () => {
  it('exige un token Bearer', async () => {
    const response = await request(createTestApp())
      .post('/api/v1/statistics')
      .send({ matrices: { q: [[1]] } });

    assert.equal(response.status, 401);
    assert.equal(response.body.error.code, 'unauthorized');
  });

  it('rechaza un token que no es válido', async () => {
    const response = await request(createTestApp())
      .post('/api/v1/statistics')
      .set('Authorization', 'Bearer not-a-token')
      .send({ matrices: { q: [[1]] } });

    assert.equal(response.status, 401);
  });

  it('devuelve las estadísticas de las matrices que envía la API en Go', async () => {
    const app = createTestApp();
    const token = await login(app);

    const response = await request(app)
      .post('/api/v1/statistics')
      .set('Authorization', `Bearer ${token}`)
      .send({ matrices: { q: [[1, 0], [0, -1]], r: [[2, 3], [0, 4]] } });

    assert.equal(response.status, 200);
    assert.deepEqual(response.body.statistics, {
      count: 8,
      sum: 9,
      average: 9 / 8,
      min: -1,
      max: 4,
      diagonal: { any: true, matrices: { q: true, r: false } },
    });
  });

  it('rechaza una matriz inválida', async () => {
    const app = createTestApp();
    const token = await login(app);

    const response = await request(app)
      .post('/api/v1/statistics')
      .set('Authorization', `Bearer ${token}`)
      .send({ matrices: { q: [[1, 2], [3]] } });

    assert.equal(response.status, 400);
    assert.equal(response.body.error.code, 'invalid_matrix');
  });

  it('devuelve 404 en rutas desconocidas', async () => {
    const response = await request(createTestApp()).get('/ruta-desconocida');

    assert.equal(response.status, 404);
    assert.equal(response.body.error.code, 'not_found');
  });
});
