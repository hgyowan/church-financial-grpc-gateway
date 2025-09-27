window.onload = function() {
  //<editor-fold desc="Changeable Configuration Block">

  // the following lines will be replaced by docker/configurator, when it runs in a docker-container
  window.ui = SwaggerUIBundle({
    urls: [
      {
        url: "https://bucket.holyflows.com/public/cfm/swagger/user.swagger.json",
        name: "User"
      },
      {
        url: "https://bucket.holyflows.com/public/cfm/swagger/workspace.swagger.json",
        name: "Workspace"
      },
      {
        url: "https://bucket.holyflows.com/public/cfm/swagger/member.swagger.json",
        name: "Member"
      },
      {
        url: "https://bucket.holyflows.com/public/cfm/swagger/transaction.swagger.json",
        name: "Transaction"
      },
    ],
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl,
    ],
    layout: "StandaloneLayout",
  });

  //</editor-fold>
};
