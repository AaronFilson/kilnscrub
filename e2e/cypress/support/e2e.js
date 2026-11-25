// Cypress custom commands for kilnscrub UI end-to-end decision-tree verification.

Cypress.Commands.add("interceptEvaluate", () => {
  cy.intercept("POST", "/api/studio/evaluate").as("evalStudio");
});

Cypress.Commands.add("waitForEvalOK", () => {
  cy.wait("@evalStudio").then((interception) => {
    expect(interception.response?.statusCode, "POST /api/studio/evaluate status").to.eq(200);
    expect(interception.response?.body).to.have.property("design");
    expect(interception.response?.body).to.have.property("cots");
  });
  // Ensure no validation error notice is rendered in the UI.
  cy.contains("Validation / API Notice:").should("not.exist");
});
