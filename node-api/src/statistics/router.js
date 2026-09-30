import { Router } from 'express';

import { computeStatistics, normalizeMatrices } from './service.js';

export function createStatisticsRouter() {
  const router = Router();

  router.post('/statistics', (req, res, next) => {
    try {
      const matrices = normalizeMatrices(req.body ?? {});
      return res.json({ statistics: computeStatistics(matrices) });
    } catch (error) {
      return next(error);
    }
  });

  return router;
}
