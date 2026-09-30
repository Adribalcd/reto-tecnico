export class ApiError extends Error {
  constructor(status, code, message) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

export function notFound(req, res, next) {
  next(new ApiError(404, 'not_found', `la ruta ${req.method} ${req.originalUrl} no existe`));
}

export function errorHandler(error, req, res, next) {
  if (res.headersSent) {
    return next(error);
  }

  if (error instanceof ApiError) {
    return res.status(error.status).json({ error: { code: error.code, message: error.message } });
  }

  if (error?.type === 'entity.parse.failed' || error instanceof SyntaxError) {
    return res.status(400).json({ error: { code: 'invalid_body', message: 'el cuerpo debe ser JSON válido' } });
  }

  console.error(`[node-api] error no controlado en ${req.method} ${req.originalUrl}:`, error);
  return res.status(500).json({ error: { code: 'internal_error', message: 'error inesperado del servidor' } });
}
