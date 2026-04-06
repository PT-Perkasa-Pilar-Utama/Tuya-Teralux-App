#!/bin/bash
# generate_openapi.sh - Orchestrates Swagger 2.0 to OpenAPI 3.1 conversion
#
# This script:
# 1. Reads Swagger 2.0 JSON from docs/swagger/swagger.json
# 2. Converts to OpenAPI 3.1.0 using custom tool
# 3. Outputs to docs/openapi/ directory

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/../../backend" && pwd)"
SWAGGER_JSON="$BACKEND_DIR/docs/swagger/swagger.json"
OPENAPI_DIR="$BACKEND_DIR/docs/openapi"
CONVERTER_BIN="$BACKEND_DIR/bin/openapi_convert"

echo "📍 Backend dir: $BACKEND_DIR"
echo "📍 Swagger input: $SWAGGER_JSON"
echo "📍 OpenAPI output: $OPENAPI_DIR"

# Check if swagger.json exists
if [ ! -f "$SWAGGER_JSON" ]; then
    echo "❌ Error: Swagger JSON not found at $SWAGGER_JSON"
    echo "   Run 'make swagger' first to generate Swagger 2.0 docs"
    exit 1
fi

# Check if converter binary exists
if [ ! -f "$CONVERTER_BIN" ]; then
    echo "❌ Error: OpenAPI converter not found at $CONVERTER_BIN"
    echo "   Run 'make openapi-tools' first to build the converter"
    exit 1
fi

# Create output directory
mkdir -p "$OPENAPI_DIR"

# Run conversion
echo "🔄 Running OpenAPI converter..."
"$CONVERTER_BIN" -input "$SWAGGER_JSON" -output "$OPENAPI_DIR"

# Verify output files
if [ ! -f "$OPENAPI_DIR/openapi.json" ]; then
    echo "❌ Error: Conversion failed - openapi.json not generated"
    exit 1
fi

if [ ! -f "$OPENAPI_DIR/openapi.yaml" ]; then
    echo "❌ Error: Conversion failed - openapi.yaml not generated"
    exit 1
fi

echo "✅ OpenAPI 3.1 files generated successfully:"
echo "   - $OPENAPI_DIR/openapi.json"
echo "   - $OPENAPI_DIR/openapi.yaml"

# Fix servers to use relative URL (auto-detect origin)
echo "🔧 Fixing servers to use relative URL..."
python3 -c "
import json
with open('$OPENAPI_DIR/openapi.json', 'r') as f:
    data = json.load(f)
data['servers'] = [{'url': '/', 'description': 'Current server (auto-detected)'}]
with open('$OPENAPI_DIR/openapi.json', 'w') as f:
    json.dump(data, f, indent=4)
"
echo "✅ Servers updated to use '/' (will auto-detect origin)"

# Add domain and subdomain prefixes to route summaries
echo "🏷️  Adding domain/subdomain prefixes to route summaries..."
python3 -c "
import json
import re

OPENAPI_FILE = '$OPENAPI_DIR/openapi.json'

# Domain mapping from tag
domain_map = {
    '01. Tuya': 'Tuya',
    '02. Terminal': 'Terminal',
    '03. Scenes': 'Scenes',
    '04. Models': 'Models',
    '05. Models-v1': 'Models-v1',
    '06. Recordings': 'Recordings',
    '07. Mail': 'Mail',
    '08. Common': 'Common',
}

# Subdomain detection from path
subdomain_rules = [
    # Terminal subdomains (check more specific first)
    (r'/devices/:id/status', 'DeviceStatus'),
    (r'/devices/:id/statuses', 'DeviceStatus'),
    (r'/devices/', 'Device'),
    (r'/terminal/', 'Terminal'),

    # Models subdomains (check v1 first, then legacy)
    (r'/v1/models/rag/', 'RAG'),
    (r'/v1/models/whisper/', 'Whisper'),
    (r'/v1/models/pipeline/', 'Pipeline'),
    (r'/models/rag/', 'RAG'),
    (r'/models/whisper/', 'Whisper'),
    (r'/models/pipeline/', 'Pipeline'),
]

with open(OPENAPI_FILE, 'r') as f:
    data = json.load(f)

for path, methods in data.get('paths', {}).items():
    for method, operation in methods.items():
        if method not in ['get', 'post', 'put', 'delete', 'patch']:
            continue

        tags = operation.get('tags', [])
        if not tags:
            continue

        tag = tags[0]
        domain = domain_map.get(tag)

        if not domain:
            continue

        # Detect subdomain from path
        subdomain = None
        for pattern, subdomain_name in subdomain_rules:
            if re.search(pattern, path):
                subdomain = subdomain_name
                break

        # Build prefix
        if subdomain:
            prefix = f'[{domain}] [{subdomain}]'
        else:
            prefix = f'[{domain}]'

        # Update summary
        summary = operation.get('summary', '')
        if summary and not summary.startswith(prefix):
            operation['summary'] = f'{prefix} {summary}'

# Fix BearerAuth security scheme to use proper HTTP Bearer format
if 'components' in data and 'securitySchemes' in data['components']:
    if 'BearerAuth' in data['components']['securitySchemes']:
        # Convert from apiKey to http bearer scheme
        data['components']['securitySchemes']['BearerAuth'] = {
            'type': 'http',
            'scheme': 'bearer',
            'bearerFormat': 'JWT',
            'description': 'Enter JWT token only (without \"Bearer \" prefix) - Swagger UI will add it automatically'
        }

with open(OPENAPI_FILE, 'w') as f:
    json.dump(data, f, indent=4)

print('✅ Route summaries updated with domain/subdomain prefixes')
print('✅ BearerAuth security scheme fixed to use HTTP Bearer format')
"
