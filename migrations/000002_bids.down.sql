DROP TRIGGER trg_update_bid_version ON bid;
DROP FUNCTION update_bid_version();
DROP TABLE IF EXISTS bid_history;
DROP TABLE IF EXISTS bid;
DROP TYPE author_type;
DROP TYPE bid_status;