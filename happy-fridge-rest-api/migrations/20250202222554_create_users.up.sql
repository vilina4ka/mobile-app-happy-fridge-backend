-- Create the users table
CREATE TABLE users (
                       id bigserial NOT NULL PRIMARY KEY,
                       email varchar(255) NOT NULL UNIQUE,
                       encrypted_password varchar(255) NOT NULL,
                       fridge_number varchar(6) NOT NULL UNIQUE
);

-- Create the function to generate the fridge_number
CREATE OR REPLACE FUNCTION generate_fridge_number()
    RETURNS trigger AS $$
DECLARE
    result varchar(6);
    chars varchar := 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
    i integer;
BEGIN
    result := '';
    FOR i IN 1..6 LOOP
            result := result || substr(chars, floor(random() * length(chars) + 1)::integer, 1);
        END LOOP;
    NEW.fridge_number := result;
    RETURN NEW;
END;
$$ ;

-- Create the trigger to generate fridge_number before insert
CREATE TRIGGER set_fridge_number
    BEFORE INSERT ON users
    FOR EACH ROW
    EXECUTE FUNCTION generate_fridge_number();