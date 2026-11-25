export default {
  e2e: {
    baseUrl: process.env.KILNSCRUB_URL || "http://127.0.0.1:8888",
    specPattern: "e2e/cypress/e2e/**/*.cy.{js,ts}",
    supportFile: "e2e/cypress/support/e2e.js",
    video: false,
    screenshotOnRunFailure: true,
    viewportWidth: 1440,
    viewportHeight: 920,
    defaultCommandTimeout: 10000,
  },
};
