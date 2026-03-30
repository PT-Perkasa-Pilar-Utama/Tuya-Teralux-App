#!/usr/bin/env python3
"""
fix_openapi.py - Fix missing requestBody schemas in OpenAPI specification

This script:
1. Reads the OpenAPI JSON file
2. Finds endpoints with requestBody but empty schema
3. Adds appropriate schema definitions based on the endpoint type
4. Writes the fixed OpenAPI JSON file
"""

import json
import sys


def fix_openapi(input_file: str, output_file: str) -> None:
    """Fix missing requestBody schemas in OpenAPI specification."""
    
    with open(input_file, 'r') as f:
        data = json.load(f)
    
    fixes_applied = 0
    
    for path, methods in data.get('paths', {}).items():
        for method, operation in methods.items():
            if method not in ['post', 'put', 'patch']:
                continue
            
            request_body = operation.get('requestBody', {})
            if not request_body:
                continue
            
            content = request_body.get('content', {})
            json_content = content.get('application/json', {})
            schema = json_content.get('schema', {})

            # Check if content['application/json'] is empty - need to add schema
            if not json_content or (isinstance(json_content, dict) and len(json_content) == 0):
                desc = request_body.get('description', '')
                schema_map = {
                    'Updated device data': 'dtos.UpdateDeviceRequestDTO',
                    'Updated status data': 'dtos.UpdateDeviceStatusRequestDTO',
                    'Updated terminal data': 'dtos.UpdateTerminalRequestDTO',
                    'Updated scene configuration': 'dtos.UpdateSceneRequestDTO',
                    'Device registration data': 'dtos.CreateDeviceRequestDTO',
                    'Terminal registration data': 'dtos.CreateTerminalRequestDTO',
                    'Scene configuration': 'dtos.CreateSceneRequestDTO',
                    'IR Command Payload': 'dtos.TuyaIRACCommandDTO',
                    'Command Payload': 'dtos.TuyaCommandDTO',
                    'Translation request': 'dtos.RAGRequestDTO',
                    'Prompt Request': 'dtos.RAGRawPromptRequestDTO',
                    'Notification details': 'dtos.NotificationPublishRequest',
                    'Mail Request': 'dtos.MailSendRequestDTO',
                }
                schema_name = schema_map.get(desc)
                if schema_name and schema_name in data.get('components', {}).get('schemas', {}):
                    request_body['content']['application/json'] = {"schema": {"$ref": f"#/components/schemas/{schema_name}"}}
                    fixes_applied += 1
                    continue

            # Check if schema has $ref - if yes, skip
            if isinstance(schema, dict) and '$ref' in schema:
                continue

            # Check if schema is empty or is inline object that should be $ref
            # If schema is inline object with same name pattern, replace with $ref
            if isinstance(schema, dict) and 'type' in schema and schema.get('type') == 'object':
                # Try to find matching schema in components based on description
                desc = request_body.get('description', '')
                # Map description to schema names
                schema_map = {
                    'Updated device data': 'dtos.UpdateDeviceRequestDTO',
                    'Updated status data': 'dtos.UpdateDeviceStatusRequestDTO',
                    'Updated terminal data': 'dtos.UpdateTerminalRequestDTO',
                    'Updated scene configuration': 'dtos.UpdateSceneRequestDTO',
                    'Device registration data': 'dtos.CreateDeviceRequestDTO',
                    'Terminal registration data': 'dtos.CreateTerminalRequestDTO',
                    'Scene configuration': 'dtos.CreateSceneRequestDTO',
                    'IR Command Payload': 'dtos.TuyaIRACCommandDTO',
                    'Command Payload': 'dtos.TuyaCommandDTO',
                    'Translation request': 'dtos.RAGRequestDTO',
                    'Prompt Request': 'dtos.RAGRawPromptRequestDTO',
                    'Notification details': 'dtos.NotificationPublishRequest',
                    'Mail Request': 'dtos.MailSendRequestDTO',
                    'RAG Chat Request': 'dtos.RAGChatRequestDTO',
                    'RAG Control Request': 'dtos.RAGControlRequestDTO',
                    'RAG Summary Request': 'dtos.RAGSummaryRequestDTO',
                    'Upload session configuration': 'sensio_domain_models-v1_whisper_dtos.CreateUploadSessionRequest',
                }
                schema_name = schema_map.get(desc)
                if schema_name and schema_name in data.get('components', {}).get('schemas', {}):
                    # Replace inline schema with $ref
                    json_content['schema'] = {"$ref": f"#/components/schemas/{schema_name}"}
                    fixes_applied += 1
                    continue
                
                # Fallback: Match based on keywords in description
                keywords = desc.lower().replace('updated', '').replace('request', '').replace('data', '').replace('configuration', '').replace('payload', '').strip().split()
                for comp_name in data.get('components', {}).get('schemas', {}).keys():
                    comp_lower = comp_name.lower()
                    # Check if all keywords appear in component name
                    if all(kw in comp_lower for kw in keywords if len(kw) > 2):
                        json_content['schema'] = {"$ref": f"#/components/schemas/{comp_name}"}
                        fixes_applied += 1
                        break

            # Check if schema is empty (missing) - skip if already has $ref or type
            if not schema or (isinstance(schema, dict) and len(schema) == 0):
                # Determine appropriate schema based on endpoint
                if '/api/devices' in path and method == 'post':
                    schema = {"$ref": "#/components/schemas/dtos.CreateDeviceRequestDTO"}
                elif '/api/terminal' in path and method == 'post':
                    schema = {"$ref": "#/components/schemas/dtos.CreateTerminalRequestDTO"}
                elif '/api/notification/publish' in path and method == 'post':
                    schema = {"$ref": "#/components/schemas/dtos.NotificationPublishRequest"}
                elif '/api/mail/send' in path:
                    schema = {"$ref": "#/components/schemas/dtos.MailSendRequestDTO"}
                elif '/api/models/rag' in path and '/chat' in path:
                    schema = {"$ref": "#/components/schemas/dtos.RAGChatRequestDTO"}
                elif '/api/models/rag' in path and '/control' in path:
                    schema = {"$ref": "#/components/schemas/dtos.RAGControlRequestDTO"}
                elif '/api/models/rag' in path and '/summary' in path:
                    schema = {"$ref": "#/components/schemas/dtos.RAGSummaryRequestDTO"}
                elif '/api/models/gemini' in path or '/api/models/groq' in path or \
                     '/api/models/llama' in path or '/api/models/openai' in path:
                    schema = {"$ref": "#/components/schemas/dtos.RAGRawPromptRequestDTO"}
                elif '/api/scene' in path and method == 'post':
                    schema = {"$ref": "#/components/schemas/dtos.CreateSceneRequestDTO"}
                elif '/api/scene' in path and method in ['put', 'patch']:
                    schema = {"$ref": "#/components/schemas/dtos.UpdateSceneRequestDTO"}
                elif '/api/tuya/device' in path and method == 'post':
                    schema = {"$ref": "#/components/schemas/dtos.TuyaCommandDTO"}
                elif '/api/tuya/irac' in path and method == 'post':
                    schema = {"$ref": "#/components/schemas/dtos.TuyaIRACCommandDTO"}
                elif '/api/v1/models/whisper' in path and '/upload' in path:
                    schema = {"$ref": "#/components/schemas/sensio_domain_models-v1_whisper_dtos.UploadChunkAckDTO"}
                elif '/api/v1/models/rag' in path:
                    schema = {"$ref": "#/components/schemas/services.RAGRequest"}
                else:
                    # Default: use object type with description
                    schema = {
                        "type": "object",
                        "description": request_body.get('description', 'Request body')
                    }
                
                # Apply the fix
                if 'content' not in request_body:
                    request_body['content'] = {}
                if 'application/json' not in request_body['content']:
                    request_body['content']['application/json'] = {}
                request_body['content']['application/json']['schema'] = schema
                fixes_applied += 1
    
    # Write the fixed OpenAPI spec
    with open(output_file, 'w') as f:
        json.dump(data, f, indent=4)
    
    print(f"✅ Fixed {fixes_applied} missing requestBody schema(s)")


if __name__ == '__main__':
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} <input_openapi.json> <output_openapi.json>")
        sys.exit(1)
    
    fix_openapi(sys.argv[1], sys.argv[2])
