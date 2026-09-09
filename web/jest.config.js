export default {
  preset: 'ts-jest',
  moduleFileExtensions: [
    'js',
    'ts',
    'json',
    'vue'
  ],
  transform: {
    '^.+\\.ts$': 'ts-jest',
    '^.+\\.js$': 'babel-jest',
    '^.+\\.vue$': '@vue/vue3-jest'
  },
  testEnvironment: 'jsdom',
  testPathIgnorePatterns: [
    '<rootDir>/e2e/'
  ],
  transformIgnorePatterns: [
    '/node_modules/(?!(element-plus|@element-plus|@vueuse|@popperjs|vue-i18n|@intlify)/)'
  ],
  setupFilesAfterEnv: ['./tests/setup-after-env.js'],
  testEnvironmentOptions: {
    customExportConditions: ['node', 'node-addons']
  }
}
