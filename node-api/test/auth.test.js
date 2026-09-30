import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import jwt from 'jsonwebtoken';
import request from 'supertest';

import { signToken, verifyToken } from '../src/auth/tokens.js';
import { createTestApp, testConfig } from './helpers.js';

describe('utilidades de JWT', () => {
  it('firma un token que su propio verificador acepta', () => {
    const token = signToken(testConfig, 'admin');
    const claims = verifyToken(testConfig, token);

    assert.equal(claims.sub, 'admin');
    assert.equal(claims.iss, testConfig.jwtIssuer);
  });

  it('rechaza un token firmado con otro secreto', () => {
    const token = jwt.sign({}, 'otro-secreto', {
      subject: 'admin',
      issuer: testConfig.jwtIssuer,
      algorithm: 'HS256',
      expiresIn: '1h',
    });

    assert.throws(() => verifyToken(testConfig, token));
  });

  it('rechaza un token expirado', async () => {
    const token = jwt.sign({}, testConfig.jwtSecret, {
      subject: 'admin',
      issuer: testConfig.jwtIssuer,
      algorithm: 'HS256',
      expiresIn: '-1s',
    });

    const response = await request(createTestApp())
      .post('/api/v1/statistics')
      .set('Authorization', `Bearer ${token}`)
      .send({ matrices: { q: [[1]] } });

    assert.equal(response.status, 401);
    assert.equal(response.body.error.code, 'unauthorized');
  });
});
