module.exports = {
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
    '/node_modules/(?!(element-plus|@element-plus|@vueuse|@popperjs)/)'
  ],
  setupFilesAfterEnv: ['./tests/setup-after-env.js'],
  testEnvironmentOptions: {
    customExportConditions: ['node', 'node-addons']
  }
}
