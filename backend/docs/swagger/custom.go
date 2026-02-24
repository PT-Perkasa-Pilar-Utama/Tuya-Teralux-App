package swagger

const CustomSwaggerHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Teralux API - Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3/swagger-ui.css" >
  <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@3/favicon-32x32.png" sizes="32x32" />
  <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@3/favicon-16x16.png" sizes="16x16" />
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin:0; background: #fafafa; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@3/swagger-ui-bundle.js"> </script>
  <script src="https://unpkg.com/swagger-ui-dist@3/swagger-ui-standalone-preset.js"> </script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "/swagger/doc.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout",
        requestInterceptor: (req) => {
          return req;
        },
        responseInterceptor: (res) => {
          // Auto-fill bearer token when calling /api/tuya/auth
          if (res.url && res.url.includes('/api/tuya/auth') && res.status === 200) {
            try {
              const data = JSON.parse(res.text);
              if (data.status && data.data && data.data.access_token) {
                const token = data.data.access_token;
                // Auto-fill bearer token
                ui.preauthorizeApiKey('BearerAuth', 'Bearer ' + token);
                console.log('✅ Bearer token auto-filled successfully!');
              }
            } catch (e) {
              console.error('Failed to parse auth response:', e);
            }
          }
          return res;
        }
      })
      window.ui = ui
    }
  </script>
</body>
</html>`
