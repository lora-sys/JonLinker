import { antfu } from '@antfu/eslint-config'

export default antfu({
  react: true,
  nextjs: true,
  typescript: {
    overrides: {
      'ts/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      'ts/prefer-ts-expect-error': 'off',
    },
  },
  formatters: {
    css: true,
    html: true,
  },
  ignores: [
    'dist/',
    'build/',
    'node_modules/',
    '.next/',
    'next-env.d.ts',
    'test-results/',
  ],
}, {
  rules: {
    'no-console': 'off',
    'react-refresh/only-export-components': 'off',
    'unused-imports/no-unused-imports': 'error',
    'perfectionist/sort-imports': ['error', {
      type: 'natural',
      order: 'asc',
      internalPattern: ['^@/'],
    }],
    'jsonc/sort-keys': 'off',
    'node/prefer-global/process': 'off',
    'node/prefer-global/buffer': 'off',
    'react/no-array-index-key': 'warn',
    'react/set-state-in-effect': 'off',
    'react/purity': 'off',
    'react/no-forward-ref': 'off',
    'react/no-context-provider': 'off',
    'react/no-use-context': 'off',
    'react/use-state': 'off',
    'eslint-comments/no-unlimited-disable': 'off',
    'react/naming-convention-ref-name': 'off',
    'react/no-unstable-context-value': 'off',
  },
})
