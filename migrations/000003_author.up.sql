CREATE OR REPLACE FUNCTION check_author_integrity()
RETURNS TRIGGER AS $$
BEGIN
    IF (NEW.author_type = 'Organization') THEN
        IF NOT EXISTS (SELECT 1 FROM organization WHERE id = NEW.author_id) THEN
            RAISE EXCEPTION 'Invalid author_id % for author_type %', NEW.author_id, NEW.author_type;
        END IF;
    ELSE
        IF NOT EXISTS (SELECT 1 FROM employee WHERE id = NEW.author_id) THEN
            RAISE EXCEPTION 'Invalid author_id % for author_type %', NEW.author_id, NEW.author_type;
        END IF;
    END IF;
    RETURN NEW;
END; $$ LANGUAGE plpgsql;

CREATE TRIGGER author_integrity_trigger
    BEFORE INSERT OR UPDATE ON bid FOR EACH ROW EXECUTE FUNCTION check_author_integrity();