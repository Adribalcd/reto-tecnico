import cors from 'cors';
import express from 'express';

import { requireAuth } from './auth/middleware.js';
import { createAuthRouter } from './auth/router.js';
import { errorHandler, notFound } from './http/errors.js';
import { createStatisticsRouter } from './statistics/router.js';

export function createApp({ config }) {
  const app = express();

  app.disable('x-powered-by');
  app.use(cors());
  app.use(express.json({ limit: '2mb' }));

  app.get('/health', (req, res) => res.json({ status: 'ok', service: 'node-api' }));

  app.use('/auth', createAuthRouter({ config }));
  app.use('/api/v1', requireAuth(config), createStatisticsRouter());

  app.use(notFound);
  app.use(errorHandler);

  return app;
}
