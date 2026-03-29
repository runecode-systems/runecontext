# Verification

## Notes
Run go test ./internal/cli -run Change and just test. Cover verified-state success when verification_status is supplied, rejection when verified still implies pending verification_status, non-terminal updates that leave promotion_assessment untouched, and backward/terminal transition rejections.

## Planned Checks
- `just test`

## Close Gate
Use the repository's standard verification flow before closing this change.
