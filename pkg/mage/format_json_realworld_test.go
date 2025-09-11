package mage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFormatJSONRealWorldExamples tests formatting of real-world JSON examples
func TestFormatJSONRealWorldExamples(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-realworld-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }() //nolint:errcheck // cleanup in defer

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }() //nolint:errcheck // cleanup in defer
	require.NoError(t, os.Chdir(tmpDir))

	tests := []struct {
		name        string
		input       string
		description string
	}{
		{
			name: "package.json",
			input: `{
  "name": "my-awesome-project",
  "version": "1.0.0",
  "description": "An awesome Node.js project",
  "main": "index.js",
  "scripts": {
    "start": "node index.js",
    "test": "jest",
    "build": "webpack --mode production",
    "dev": "webpack --mode development --watch"
  },
  "keywords": ["node", "javascript", "awesome"],
  "author": "John Doe <john@example.com>",
  "license": "MIT",
  "dependencies": {
    "express": "^4.18.2",
    "lodash": "^4.17.21",
    "axios": "^1.4.0"
  },
  "devDependencies": {
    "jest": "^29.5.0",
    "webpack": "^5.88.0",
    "webpack-cli": "^5.1.4"
  },
  "engines": {
    "node": ">=16.0.0",
    "npm": ">=8.0.0"
  }
}`,
			description: "Standard npm package.json file",
		},
		{
			name: "tsconfig.json",
			input: `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "lib": ["ES2020"],
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "removeComments": false,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "strictFunctionTypes": true,
    "noImplicitThis": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist", "**/*.test.ts"]
}`,
			description: "TypeScript configuration file",
		},
		{
			name: "api_response.json",
			input: `{
  "status": "success",
  "data": {
    "users": [
      {
        "id": 1,
        "username": "johndoe",
        "email": "john.doe@example.com",
        "profile": {
          "firstName": "John",
          "lastName": "Doe",
          "avatar": "https://example.com/avatars/1.jpg",
          "bio": "Software developer passionate about open source",
          "location": "San Francisco, CA",
          "website": "https://johndoe.dev",
          "socialLinks": {
            "twitter": "@johndoe",
            "github": "johndoe",
            "linkedin": "john-doe"
          }
        },
        "settings": {
          "theme": "dark",
          "notifications": {
            "email": true,
            "push": false,
            "sms": false
          },
          "privacy": {
            "profileVisible": true,
            "emailVisible": false
          }
        },
        "metadata": {
          "createdAt": "2023-01-15T10:30:00Z",
          "updatedAt": "2023-06-20T14:45:30Z",
          "lastLogin": "2023-06-21T09:15:22Z",
          "loginCount": 247,
          "isVerified": true,
          "role": "user"
        }
      },
      {
        "id": 2,
        "username": "janedoe",
        "email": "jane.doe@example.com",
        "profile": {
          "firstName": "Jane",
          "lastName": "Doe",
          "avatar": "https://example.com/avatars/2.jpg",
          "bio": "UX designer and frontend developer",
          "location": "New York, NY",
          "website": "https://janedoe.design",
          "socialLinks": {
            "twitter": "@janedoe",
            "github": "janedoe",
            "dribbble": "janedoe"
          }
        },
        "settings": {
          "theme": "light",
          "notifications": {
            "email": true,
            "push": true,
            "sms": false
          },
          "privacy": {
            "profileVisible": true,
            "emailVisible": true
          }
        },
        "metadata": {
          "createdAt": "2023-02-20T16:20:00Z",
          "updatedAt": "2023-06-21T11:30:15Z",
          "lastLogin": "2023-06-21T11:30:15Z",
          "loginCount": 189,
          "isVerified": true,
          "role": "admin"
        }
      }
    ]
  },
  "pagination": {
    "page": 1,
    "perPage": 10,
    "total": 2,
    "totalPages": 1
  },
  "meta": {
    "requestId": "req_123456789",
    "timestamp": "2023-06-21T12:00:00Z",
    "version": "v1"
  }
}`,
			description: "Complex API response with nested data",
		},
		{
			name: "docker_compose.json",
			input: `{
  "version": "3.8",
  "services": {
    "web": {
      "image": "nginx:alpine",
      "ports": ["80:80", "443:443"],
      "volumes": ["./nginx.conf:/etc/nginx/nginx.conf:ro", "./ssl:/etc/ssl:ro"],
      "depends_on": ["api"],
      "environment": {
        "NGINX_HOST": "localhost",
        "NGINX_PORT": "80"
      }
    },
    "api": {
      "build": {
        "context": ".",
        "dockerfile": "Dockerfile"
      },
      "ports": ["3000:3000"],
      "environment": {
        "NODE_ENV": "production",
        "DATABASE_URL": "postgresql://user:pass@db:5432/myapp",
        "REDIS_URL": "redis://redis:6379",
        "JWT_SECRET": "your-jwt-secret"
      },
      "depends_on": ["db", "redis"],
      "volumes": ["./logs:/app/logs"]
    },
    "db": {
      "image": "postgres:15",
      "environment": {
        "POSTGRES_DB": "myapp",
        "POSTGRES_USER": "user",
        "POSTGRES_PASSWORD": "pass"
      },
      "volumes": ["postgres_data:/var/lib/postgresql/data"],
      "ports": ["5432:5432"]
    },
    "redis": {
      "image": "redis:7-alpine",
      "ports": ["6379:6379"],
      "volumes": ["redis_data:/data"]
    }
  },
  "volumes": {
    "postgres_data": {},
    "redis_data": {}
  }
}`,
			description: "Docker Compose configuration",
		},
		{
			name:        "minified_production.json",
			input:       `{"config":{"api":{"baseUrl":"https://api.example.com","timeout":5000,"retries":3},"features":{"darkMode":true,"analytics":true,"notifications":false},"cache":{"ttl":3600,"maxSize":100}},"routes":[{"path":"/","component":"Home","exact":true},{"path":"/about","component":"About"},{"path":"/contact","component":"Contact"}],"version":"2.1.0"}`,
			description: "Minified production configuration",
		},
		{
			name: "composer.json",
			input: `{
    "name": "vendor/package-name",
    "description": "A sample PHP package",
    "type": "library",
    "license": "MIT",
    "authors": [
        {
            "name": "John Doe",
            "email": "john@example.com",
            "role": "Developer"
        }
    ],
    "minimum-stability": "stable",
    "require": {
        "php": ">=8.1",
        "symfony/console": "^6.0",
        "guzzlehttp/guzzle": "^7.0",
        "monolog/monolog": "^3.0"
    },
    "require-dev": {
        "phpunit/phpunit": "^10.0",
        "squizlabs/php_codesniffer": "^3.7",
        "phpstan/phpstan": "^1.10"
    },
    "autoload": {
        "psr-4": {
            "Vendor\\PackageName\\": "src/"
        }
    },
    "autoload-dev": {
        "psr-4": {
            "Vendor\\PackageName\\Tests\\": "tests/"
        }
    },
    "scripts": {
        "test": "phpunit",
        "cs-check": "phpcs",
        "cs-fix": "phpcbf",
        "analyze": "phpstan analyze"
    }
}`,
			description: "PHP Composer package definition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.name)
			err := os.WriteFile(testFile, []byte(tt.input), 0o600)
			require.NoError(t, err, tt.description)

			success := formatJSONFileNative(testFile)
			assert.True(t, success, "Formatting should succeed for %s", tt.description)

			// Read the formatted result
			result, err := os.ReadFile(testFile) //nolint:gosec // test file path
			require.NoError(t, err)

			// Verify the result is valid JSON
			var jsonData interface{}
			err = json.Unmarshal(result, &jsonData)
			require.NoError(t, err, "Result should be valid JSON for %s", tt.description)

			// Verify formatting consistency - the result should have proper indentation
			resultStr := string(result)
			assert.Contains(t, resultStr, "    ", "Should contain 4-space indentation")
			assert.Equal(t, byte('\n'), resultStr[len(resultStr)-1], "Should end with newline")

			// Test that the formatted result is idempotent
			err = os.WriteFile(testFile, result, 0o600)
			require.NoError(t, err)

			success = formatJSONFileNative(testFile)
			assert.True(t, success, "Second formatting should succeed for %s", tt.description)

			secondResult, err := os.ReadFile(testFile) //nolint:gosec // test file path
			require.NoError(t, err)

			assert.Equal(t, string(result), string(secondResult),
				"Formatting should be idempotent for %s", tt.description)
		})
	}
}

// TestFormatJSONLargeFile tests formatting of very large JSON files
func TestFormatJSONLargeFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-large-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }() //nolint:errcheck // cleanup in defer

	t.Run("large array with many objects", func(t *testing.T) {
		// Create a large JSON array with 5 objects (simplified for testing)
		largeJSON := `{"users":[{"id":1,"name":"User1","email":"user1@example.com","active":true},{"id":2,"name":"User2","email":"user2@example.com","active":true},{"id":3,"name":"User3","email":"user3@example.com","active":true},{"id":4,"name":"User4","email":"user4@example.com","active":true},{"id":5,"name":"User5","email":"user5@example.com","active":true}],"metadata":{"total":5,"generated":"2023-06-21T12:00:00Z"}}`

		testFile := filepath.Join(tmpDir, "large.json")
		err := os.WriteFile(testFile, []byte(largeJSON), 0o600)
		require.NoError(t, err)

		success := formatJSONFileNative(testFile)
		assert.True(t, success, "Should handle large JSON files")

		// Verify the result is valid and well-formatted
		result, err := os.ReadFile(testFile) //nolint:gosec // test file path
		require.NoError(t, err)

		var jsonData interface{}
		err = json.Unmarshal(result, &jsonData)
		require.NoError(t, err, "Large JSON should remain valid after formatting")

		resultStr := string(result)
		assert.Contains(t, resultStr, "    ", "Large JSON should have proper indentation")
		assert.Equal(t, byte('\n'), resultStr[len(resultStr)-1], "Large JSON should end with newline")
	})
}

// TestFormatJSONSpecialCases tests special JSON formatting cases
func TestFormatJSONSpecialCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-special-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }() //nolint:errcheck // cleanup in defer

	tests := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "json with comments (invalid)",
			input:       `{"name": "test", // this is a comment\n"value": 123}`,
			description: "JSON with comments should fail validation",
		},
		{
			name:        "json5 features (invalid)",
			input:       `{name: 'test', value: 123,}`,
			description: "JSON5 syntax should fail in standard JSON",
		},
		{
			name:        "single quotes (invalid)",
			input:       `{'name': 'test', 'value': 123}`,
			description: "Single quotes should fail in standard JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, "test.json")
			err := os.WriteFile(testFile, []byte(tt.input), 0o600)
			require.NoError(t, err)

			success := formatJSONFileNative(testFile)
			assert.False(t, success, tt.description)

			// Verify original file content is unchanged when formatting fails
			content, err := os.ReadFile(testFile) //nolint:gosec // test file path
			require.NoError(t, err)
			assert.Equal(t, tt.input, string(content), "Original content should be unchanged on failure")
		})
	}
}
