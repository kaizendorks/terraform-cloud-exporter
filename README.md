# Terraform Cloud/Enterprise Exporter
>Prometheus exporter for Terraform Cloud/Enterprise metrics.

[![Go Test](https://github.com/kaizendorks/terraform-cloud-exporter/workflows/Go%20Test/badge.svg)](https://github.com//kaizendorks/terraform-cloud-exporter/actions?workflow=Go%20Test)
[![GitHub release](https://img.shields.io/github/release/kaizendorks/terraform-cloud-exporter.svg)](https://github.com//kaizendorks/terraform-cloud-exporter/releases/latest)
[![Docker Release](https://github.com/kaizendorks/terraform-cloud-exporter/workflows/Docker%20Release/badge.svg)](https://github.com//kaizendorks/terraform-cloud-exporter/actions?workflow=Docker%20Release)


## Description
| ![Sample Dashboard](workspace-dasbhoard.png?raw=true "Sample Dashboard") |
|:--:|
| *Sample dashboard available in* `grafana/dashboards/general.json` |

## Usage
1. Create API Token:
    * For Terraform Cloud open: `https://app.terraform.io/app/settings/tokens?source=terraform-login`
    * For Terraform Enterprise see: `https://www.terraform.io/docs/cloud/users-teams-organizations/api-tokens.html`
1. Save the token to a file or as the environment variable `TF_API_TOKEN`.

#### Native
1. `go get -u github.com/kaizendorks/terraform-cloud-exporter`
1. `terraform-cloud-exporter -o <YourOrg> --api-token-file=/path/to/file`

#### Docker

        docker run -it --rm \
            -p 9100:9100 \
            -e TF_API_TOKEN=<YourToken> \
        terraform-cloud-exporter -o <YourOrg>

### Full list of Flags

        -h, --help                                     Show context-sensitive help.
        -o, --organization=STRING                      Name of the Organization to scrape from ($TF_ORGANIZATION).
        -t, --api-token=STRING                         User token for autheticating with the API ($TF_API_TOKEN).
            --api-token-file=/path/to/file             File containing user token for autheticating with the API.
            --api-address=https://app.terraform.io/    Terraform API address to scrape metrics from.
            --api-insecure-skip-verify                 Accept any certificate presented by the API.
            --listen-address="0.0.0.0:9100"            Address to listen on for web interface and telemetry.
            --log-level="info"                         Only log messages with the given severity or above. One of: [debug,info,warn,error]
            --log-format="logfmt"                      Output format of log messages. One of: [logfmt,json]

## Contributing
#### Dev environment
1. Create a `.env` file with your token:

        TF_API_TOKEN=<Your.atlasv1.Token>
        TF_ORGANIZATION=<YourOrg>
        TF_API_ADDRESS=<YourApiAddress - Optional: Only required for Terraform enterprise.>
1. Run the Exporter, in one of two modes:
    1. Standalone exporter: `docker-compose run --rm --service-ports --entrypoint sh exporter`
        * Run code: `go run main.go`
        * View metricts: `curl localhost:9100/metrics`
    1. Full Prometheus stack: `docker-compose up`
        * Open Grafana: `http://localhost:3000/`
        * Clean up: `docker-compode down`

#### Go tests
1. Apply style guides: `go fmt ./...`
1. Static analysis: `go vet ./...`
1. Run tests: `go test -v ./... -coverprofile cover.out`
1. Examine code coverage: `go tool cover -func=cover.out`

#### Prod Image tests
1. Todo: Clean this up....
1. Build prod image:

        docker build --target prod \
            --build-arg tag=localv \
            --build-arg sha=locals \
        -t terraform-cloud-exporter .
1. Run prod image

        docker run -it --rm \
            --env-file .env \
            -p 9100:9100 \
        terraform-cloud-exporter \
            --log-level debug
1. Todo: Add dgoss tests
