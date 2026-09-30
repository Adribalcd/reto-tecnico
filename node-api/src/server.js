import { createApp } from './app.js';
import { loadConfig } from './config.js';

const config = loadConfig();

if (config.users.size === 0) {
  console.warn('[node-api] no hay usuarios configurados; define AUTH_USERS para permitir el login');
}

const app = createApp({ config });

const server = app.listen(config.port, () => {
  console.log(`[node-api] escuchando en el puerto ${config.port}`);
});

for (const signal of ['SIGINT', 'SIGTERM']) {
  process.on(signal, () => {
    console.log(`[node-api] señal ${signal} recibida, apagando`);
    server.close(() => process.exit(0));
  });
}
