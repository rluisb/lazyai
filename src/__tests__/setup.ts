// Skip post-install validation in tests to avoid invoking the real `opencode`
// CLI (which takes ~1.5s per call and would make the wizard-heavy test suite
// take 10+ minutes). Validation is covered in isolation by
// opencode-validate.test.ts, which uses an injectable runner and does not set
// this flag.
process.env.AI_SETUP_SKIP_VALIDATION = '1'
