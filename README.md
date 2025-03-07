# Email Verifier

A robust email verification tool that helps validate email addresses through various checks including syntax validation, DNS records, and SMTP verification.

## Features

- **Email Validation**: Verify email addresses for syntax, domain validity, and deliverability
- **Batch Processing**: Process large lists of emails efficiently with multi-threading
- **Multiple File Formats**: Support for CSV and XLSX input/output formats
- **Case-Insensitive Handling**: Process email addresses and column names regardless of case
- **Comprehensive Verification**: Check MX records, SMTP responses, and more
- **API Mode**: Expose verification functionality as a REST API
- **Google Sheets Integration**: Verify emails directly in Google Sheets

## Installation

### Prerequisites

- Go 1.18 or higher
- Internet connection for DNS and SMTP checks
- Port 25 open for SMTP verification (or use a proxy)

### Building from Source

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/email-verifier.git
   cd email-verifier
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. Build the application:
   ```
   go build -o email_verifier main.go
   ```

## Configuration

Create a `config.yaml` file with the following settings:

```yaml
# Input/Output Configuration
input_file: "leads.csv"
input_type: "csv"  # csv or xlsx
output_file: "verified_leads.csv"
output_type: "csv"  # csv or xlsx

# Verification Settings
valid_threshold: 80
risky_threshold: 60
default_risky_score: 50
max_retries: 3
initial_backoff: 1s
num_workers: 10

# Scoring Weights
scoring_weights:
  has_mx_records: 20
  reachable_yes: 40
  reachable_unknown: 20
  role_account: -10
  free_provider: -5
  suggestion: -10
```

## Usage

### Command Line Mode

Run the application with:

```
./email_verifier
```

This will process the input file specified in the config and generate an output file with verification results.

### API Mode

Run the application in API mode:

```
./cmd/api/api -port 8080
```

This will start an API server on port 8080 that exposes the following endpoints:

- `GET /health` - Health check endpoint
- `POST /verify` - Verify a single email
- `POST /batch-verify` - Verify multiple emails
- `POST /google-sheets` - Special endpoint for Google Sheets integration

#### API Examples

Verify a single email:

```bash
curl -X POST http://localhost:8080/verify \
  -H "Content-Type: application/json" \
  -d '{"email": "example@example.com"}'
```

Verify multiple emails:

```bash
curl -X POST http://localhost:8080/batch-verify \
  -H "Content-Type: application/json" \
  -d '{"emails": ["example1@example.com", "example2@example.com"]}'
```

### Google Sheets Integration

1. Open your Google Sheet
2. Go to Extensions > Apps Script
3. Copy the contents of `pkg/api/google_sheets_client.js` into the script editor
4. Update the `API_URL` constant to point to your deployed API
5. Save and reload your sheet
6. You'll see a new "Email Verifier" menu item
7. Use the menu to verify emails in your sheet

## Output Format

The verification results include:

- All original fields from the input file
- `verification_status`: One of "valid", "risky", or "invalid"
- `confidence_score`: A score from 0-100 indicating confidence in the email's validity

## Verification Logic

The tool performs several checks:

1. **Syntax Validation**: Ensures the email follows proper format
2. **MX Record Check**: Verifies the domain has valid mail exchange records
3. **SMTP Verification**: Attempts to connect to the mail server
4. **Additional Checks**: Detects disposable emails, role accounts, etc.

## Performance Optimization

- **Connection Pooling**: Reuses SMTP connections for better performance
- **Rate Limiting**: Prevents overwhelming mail servers
- **Progress Reporting**: Shows real-time verification progress

## Deploying the API

### Deploying to Render

1. Create a new Web Service on Render
2. Connect your GitHub repository
3. Set the build command: `go build -o api ./cmd/api`
4. Set the start command: `./api -port $PORT`
5. Add the environment variable: `PORT=10000`

### Deploying to Other Cloud Providers

The API can be deployed to any cloud provider that supports Go applications, including:

- Google Cloud Run
- AWS Elastic Beanstalk
- Heroku
- DigitalOcean App Platform

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 