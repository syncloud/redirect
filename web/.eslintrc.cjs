module.exports = {
  root: true,
  parserOptions: {
    ecmaVersion: 2022
  },
  extends: [
    'plugin:vue/vue3-essential',
    '@vue/standard'
  ],
  rules: {
    'no-console': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
    'no-debugger': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
    'vue/multi-word-component-names': 'off'
  },
  overrides: [
    {
      files: [
        '**/__tests__/*.{j,t}s?(x)',
        '**/tests/**/*.{j,t}s?(x)'
      ],
      env: {
        jest: true,
        node: true
      }
    },
    {
      files: ['e2e/**/*.js', '*.config.js', '*.config.cjs', '.eslintrc.cjs'],
      env: {
        node: true
      }
    }
  ]
}
