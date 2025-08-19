// src/config/env.ts
// @ts-nocheck
import { cleanEnv, str, port, bool, num } from 'envalid';

const env = cleanEnv(process.env, {  
  ENABLE_DEBUG: bool({ default: false }),

  APP_VERSION: str({ default: 'X.X.X' }),
  MAX_CONNECTIONS: num({ default: 1000 }),
  REQUEST_TIMEOUT: num({ default: 30000 }),
  NAME: str({default: 'No Name'}),
  PORT: port({ default: 5000 }),
  NODE_ENV: str({ choices: ['development', 'production', 'test'], default: 'development' }),
});

export default env;