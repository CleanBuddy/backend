-- Drop cleaner_applications table and related objects

DROP TRIGGER IF EXISTS cleaner_applications_rejection_check ON cleaner_applications;
DROP FUNCTION IF EXISTS check_rejection_reason();

DROP TRIGGER IF EXISTS cleaner_applications_updated_at ON cleaner_applications;
DROP FUNCTION IF EXISTS update_cleaner_applications_updated_at();

DROP INDEX IF EXISTS idx_cleaner_applications_pending_review;
DROP INDEX IF EXISTS idx_cleaner_applications_submitted;
DROP INDEX IF EXISTS idx_cleaner_applications_status;
DROP INDEX IF EXISTS idx_cleaner_applications_user_id;
DROP INDEX IF EXISTS idx_cleaner_applications_session_id;

DROP TABLE IF EXISTS cleaner_applications;
