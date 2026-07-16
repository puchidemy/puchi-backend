Status: DONE
Commits: 75bdbc1c1f74aadbd50b68eae97b0dd68ba49229
Tests: N/A (no tests for migration/queries)
Self-review:
- Created `003_onboarding.up.sql` — adds age_range + onboarding_completed columns to core.users, creates core.user_onboarding table
- Created `003_onboarding.down.sql` — reverses the migration cleanly
- Added 3 new queries to `users.sql`: GetUserByUsername, UpdateOnboardingInfo, UpsertUserOnboarding
- Ran `sqlc generate` successfully (exit code 0)
- Verified generated code includes:
  - `CoreUser.AgeRange string` and `CoreUser.OnboardingCompleted bool` fields
  - `CoreUserOnboarding` struct with all columns
  - All 3 new query methods in `Querier` interface
- Committed 6 files (3 modified gen files, 2 new migrations, 1 modified queries)
Concerns: none
