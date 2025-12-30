package docs

// CustomSwaggerHTML is the custom HTML template for the Swagger UI.
// It overrides the default interface to inject a custom script that automatically
// captures the access token from the login response and applies it to the "Authorize" button.
//
// Key Features:
// - Custom Styles: Applies local stylesheets.
// - Auto-Authorization: Intercepts the response from /api/tuya/auth, extracts the access_token,
//   and programmatically triggers the Swagger UI authorization action with "Bearer <token>".
// - Auto-Teralux-ID: Intercepts the response from POST /api/teralux, extracts the teralux ID,
//   stores it in localStorage, and auto-fills it in subsequent GET/UPDATE/DELETE operations.
const CustomSwaggerHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="/swagger-assets/swagger-ui.css" />
    <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16" />
    <style>
      html {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
      }

      *,
      *:before,
      *:after {
        box-sizing: inherit;
      }

      body {
        margin: 0;
        background: #fafafa;
      }
    </style>
  </head>

  <body>
    <div id="swagger-ui"></div>

    <script src="/swagger-assets/swagger-ui-bundle.js"></script>
    <script src="/swagger-assets/swagger-ui-standalone-preset.js"></script>
    <script>
      window.onload = function () {
        // Build a system
        const ui = SwaggerUIBundle({
          url: "doc.json",
          dom_id: '#swagger-ui',
          deepLinking: true,
          defaultModelsExpandDepth: -1,
          defaultModelExpandDepth: 3,
          displayRequestDuration: true,
          presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIStandalonePreset
          ],
          plugins: [
            SwaggerUIBundle.plugins.DownloadUrl
          ],
          layout: "StandaloneLayout",
          responseInterceptor: (response) => {
            // Check if this is the auth endpoint
            if (response.url && response.url.indexOf("/api/tuya/auth") > -1 && response.status === 200) {
                try {
                    console.log("Login detected, attempting to extract token...");
                    // Parse body if it isn't an object already
                    let body = response.body; 
                    if (typeof body === 'string') {
                        try {
                            body = JSON.parse(body);
                        } catch(e) {}
                    }
                    // Often response.obj is already populated by Swagger
                    const data = (body && body.data) || (response.obj && response.obj.data);

                    if (data && data.access_token) {
                        const token = data.access_token;
                        console.log("Token found:", token);
                        
                        // The security definition name in main.go is "BearerAuth"
                        const securityDefinition = "BearerAuth";
                        const bearerToken = "Bearer " + token;

                        // Trigger the authorization action
                        ui.authActions.authorize({
                            [securityDefinition]: {
                                name: securityDefinition,
                                schema: {
                                    type: "apiKey",
                                    in: "header",
                                    name: "Authorization",
                                    description: "Type 'Bearer' followed by a space and JWT token."
                                },
                                value: bearerToken
                            }
                        });
                        console.log("Token applied to Swagger UI!");
                    }
                } catch (e) {
                    console.error("Error auto-filling token:", e);
                }
            }
            
            // Check if this is the create teralux endpoint
            if (response.url && response.url.match(/\/api\/teralux$/) && response.status === 201) {
                try {
                    console.log("Teralux create detected, attempting to extract ID...");
                    let body = response.body;
                    if (typeof body === 'string') {
                        try {
                            body = JSON.parse(body);
                        } catch(e) {}
                    }
                    const data = (body && body.data) || (response.obj && response.obj.data);
                    
                    // Check for teralux_id (new standard) or id (fallback)
                    const teraluxId = (data && (data.teralux_id || data.id));
                    
                    if (teraluxId) {
                        console.log("Teralux ID found:", teraluxId);
                        
                        // Store in localStorage for use in other endpoints
                        localStorage.setItem('teralux_id', teraluxId);
                        
                        // Also display a notification
                        console.log("Teralux ID saved! Use this ID for GET/UPDATE/DELETE operations:", teraluxId);
                        
                        // Try to auto-fill immediately just in case
                        setTimeout(() => {
                            const idInputs = document.querySelectorAll('input[placeholder*="id"], input[data-param-name="id"]');
                            idInputs.forEach(input => {
                                if (input.value !== teraluxId) {
                                     // Set value securely for React
                                     const nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value").set;
                                     if (nativeInputValueSetter) {
                                        nativeInputValueSetter.call(input, teraluxId);
                                     } else {
                                        input.value = teraluxId;
                                     }
                                     input.dispatchEvent(new Event('input', { bubbles: true }));
                                     input.dispatchEvent(new Event('change', { bubbles: true }));
                                }
                            });
                        }, 500);
                    }
                } catch (e) {
                    console.error("Error auto-filling teralux ID:", e);
                }
            }
            
            return response;
          }
        });

        window.ui = ui;
      };

      // Auto-fill Teralux ID Polling Script
      setInterval(() => {
          const storedId = localStorage.getItem('teralux_id');
          if (storedId) {
             // Find all inputs that might be the ID field
             const inputs = document.querySelectorAll('input');
             inputs.forEach(input => {
                 // Check if it's an ID input (placeholder is 'id' for path params)
                 if (input.placeholder === 'id' || input.getAttribute('data-param-name') === 'id') {
                     // Verify context: ensure it's a Teralux endpoint
                     const opblock = input.closest('.opblock');
                     // opblock id usually looks like "operations-07._Teralux-get_api_teralux__id_"
                     if (opblock && opblock.id && opblock.id.toLowerCase().includes('teralux')) {
                         
                         // Check if update is needed
                         if (input.value !== storedId) {
                             // Set value securely for React
                             const nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value").set;
                             if (nativeInputValueSetter) {
                                nativeInputValueSetter.call(input, storedId);
                             } else {
                                input.value = storedId;
                             }
                             
                             input.dispatchEvent(new Event('input', { bubbles: true }));
                             input.dispatchEvent(new Event('change', { bubbles: true }));
                             
                             // console.log("Auto-updated Teralux ID input to:", storedId);
                         }
                     }
                 }
             });
          }
      }, 1000);
    </script>
  </body>
</html>
`