import js from '@eslint/js';
import globals from 'globals';

export default [
  {
    ignores: ['wasm_exec.js'],
  },
  js.configs.recommended,
  {
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: {
        ...globals.browser,
        Go: 'readonly',
        bbs: 'readonly',
      },
    },
    rules: {
      'no-console': 'off',
    },
  },
];
