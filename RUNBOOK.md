# Operations & Troubleshooting Runbook

## Common Operations

- **Deploy Application:** Use scripts in `scripts/` or CI/CD pipeline.
- **Health Check:** Run `health-check.sh` in `scripts/`.
- **Rollback:** Use `rollback.sh` in `scripts/` if needed.

## Troubleshooting

- **Frontend Issues:**
  - Check browser console for errors
  - Ensure API URL is correctly set in environment variables
- **Backend Issues:**
  - Check logs in the backend server
  - Validate database connectivity (MongoDB Atlas)
- **Infrastructure Issues:**
  - Use AWS Console and CloudWatch for diagnostics
  - Check Terraform state and logs


