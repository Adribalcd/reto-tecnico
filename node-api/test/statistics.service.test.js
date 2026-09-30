import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import { computeStatistics, normalizeMatrices } from '../src/statistics/service.js';

describe('normalizeMatrices', () => {
  it('acepta una sola matriz', () => {
    const matrices = normalizeMatrices({ matrix: [[1, 2], [3, 4]] });

    assert.deepEqual([...matrices.keys()], ['matrix']);
    assert.deepEqual(matrices.get('matrix'), [[1, 2], [3, 4]]);
  });

  it('acepta un objeto de matrices con nombre', () => {
    const matrices = normalizeMatrices({ matrices: { q: [[1, 0], [0, 1]], r: [[2]] } });

    assert.deepEqual([...matrices.keys()], ['q', 'r']);
    assert.deepEqual(matrices.get('r'), [[2]]);
  });

  it('acepta una lista de matrices', () => {
    const matrices = normalizeMatrices({ matrices: [[[1, 2]], [[3, 4]]] });

    assert.deepEqual([...matrices.keys()], ['matrix[0]', 'matrix[1]']);
  });

  it('rechaza una matriz irregular con un mensaje preciso', () => {
    assert.throws(
      () => normalizeMatrices({ matrix: [[1, 2], [3]] }),
      /matrix\[1\] tiene 1 valores, se esperaban 2/,
    );
  });

  it('rechaza valores que no son números finitos', () => {
    assert.throws(() => normalizeMatrices({ matrix: [[1, 'x']] }), /debe ser un número finito/);
    assert.throws(() => normalizeMatrices({ matrix: [[1, null]] }), /debe ser un número finito/);
  });

  it('rechaza una petición sin matriz', () => {
    assert.throws(() => normalizeMatrices({}), /debe incluir un campo matrices o matrix/);
  });
});

describe('computeStatistics', () => {
  it('resume todos los valores de las matrices', () => {
    const matrices = new Map([
      ['q', [[1, 0], [0, -1]]],
      ['r', [[2, 3], [0, 4]]],
    ]);

    const stats = computeStatistics(matrices);

    assert.equal(stats.count, 8);
    assert.equal(stats.sum, 9);
    assert.equal(stats.average, 9 / 8);
    assert.equal(stats.min, -1);
    assert.equal(stats.max, 4);
  });

  it('detecta matrices diagonales y expone la bandera global', () => {
    const matrices = new Map([
      ['q', [[1, 0], [0, -1]]],
      ['r', [[2, 3], [0, 4]]],
    ]);

    const stats = computeStatistics(matrices);

    assert.deepEqual(stats.diagonal, {
      any: true,
      matrices: { q: true, r: false },
    });
  });

  it('considera no diagonal a una matriz rectangular', () => {
    const stats = computeStatistics(new Map([['r', [[1, 0, 0], [0, 1, 0]]]]));

    assert.equal(stats.diagonal.any, false);
    assert.equal(stats.diagonal.matrices.r, false);
  });

  it('tolera el redondeo en los valores fuera de la diagonal', () => {
    const stats = computeStatistics(new Map([['q', [[1, 1e-12], [1e-12, 1]]]]));

    assert.equal(stats.diagonal.matrices.q, true);
  });
});
