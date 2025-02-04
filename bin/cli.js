#!/usr/bin/env node

import { main } from '../src/main.js';

main().catch(error => {
  console.error('Error:', error.message);
  process.exit(1);
});
