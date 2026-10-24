-- +goose Up

-- 1. messes
CREATE TABLE messes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code text UNIQUE NOT NULL,
    name text NOT NULL,
    hostel_group hostel_group NOT NULL,
    created_at timestamptz DEFAULT now()
);

-- 2. users
CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text UNIQUE NOT NULL,
    role user_role NOT NULL DEFAULT 'student',
    hostel_group hostel_group,
    diet_pref diet_pref DEFAULT 'regular',
    cvd_mode cvd_mode DEFAULT 'off',
    text_scale numeric DEFAULT 1.0 CHECK (text_scale >= 1.0 AND text_scale <= 2.0),
    theme theme_pref DEFAULT 'system',
    reduced_motion boolean DEFAULT false,
    created_at timestamptz DEFAULT now(),
    last_login_at timestamptz,
    CONSTRAINT users_email_lowercase_check CHECK (email = lower(email))
);

-- Trigger to automatically lowercase email on insert/update
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_users_email_lowercase()
RETURNS TRIGGER AS $$
BEGIN
    NEW.email = LOWER(NEW.email);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trigger_users_email_lowercase
BEFORE INSERT OR UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_users_email_lowercase();

-- 3. manager_messes
CREATE TABLE manager_messes (
    user_id uuid REFERENCES users(id) ON DELETE CASCADE,
    mess_id uuid REFERENCES messes(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, mess_id)
);

-- 4. otp_codes
CREATE TABLE otp_codes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL,
    code_hash text NOT NULL,
    expires_at timestamptz NOT NULL,
    attempts int DEFAULT 0,
    consumed_at timestamptz,
    created_at timestamptz DEFAULT now()
);

-- 5. meal_schedules
CREATE TABLE meal_schedules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    hostel_group hostel_group NOT NULL,
    meal_type meal_type NOT NULL,
    day_kind day_kind NOT NULL,
    start_time time NOT NULL,
    end_time time NOT NULL,
    CONSTRAINT meal_schedules_hostel_group_meal_type_day_kind_key UNIQUE (hostel_group, meal_type, day_kind)
);

-- 6. default_items
CREATE TABLE default_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    hostel_group hostel_group NOT NULL,
    meal_type meal_type NOT NULL,
    name text NOT NULL,
    sort_order int NOT NULL DEFAULT 0
);

-- 7. dishes
CREATE TABLE dishes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    hostel_group hostel_group NOT NULL,
    name text NOT NULL,
    created_at timestamptz DEFAULT now()
);

CREATE UNIQUE INDEX idx_dishes_hostel_group_lower_name ON dishes (hostel_group, lower(name));

-- 8. meals
CREATE TABLE meals (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    hostel_group hostel_group NOT NULL,
    service_date date NOT NULL,
    meal_type meal_type NOT NULL,
    status meal_status DEFAULT 'draft',
    published_at timestamptz,
    CONSTRAINT meals_hostel_group_service_date_meal_type_key UNIQUE (hostel_group, service_date, meal_type)
);

-- 9. meal_items
CREATE TABLE meal_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    meal_id uuid REFERENCES meals(id) ON DELETE CASCADE,
    dish_id uuid REFERENCES dishes(id),
    diet_menu diet_menu DEFAULT 'both',
    sort_order int DEFAULT 0,
    CONSTRAINT meal_items_meal_id_dish_id_key UNIQUE (meal_id, dish_id)
);

-- 10. reviews
CREATE TABLE reviews (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    meal_id uuid REFERENCES meals(id) ON DELETE CASCADE,
    user_id uuid REFERENCES users(id) ON DELETE CASCADE,
    variant diet_pref NOT NULL,
    stars smallint NOT NULL CHECK (stars >= 1 AND stars <= 5),
    comment text CHECK (char_length(comment) <= 500),
    hidden boolean DEFAULT false,
    created_at timestamptz DEFAULT now(),
    CONSTRAINT reviews_meal_id_user_id_key UNIQUE (meal_id, user_id)
);

-- 11. crowd_status
CREATE TABLE crowd_status (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    mess_id uuid REFERENCES messes(id) ON DELETE CASCADE,
    level crowd_level NOT NULL,
    updated_by uuid REFERENCES users(id) ON DELETE SET NULL,
    updated_at timestamptz DEFAULT now()
);

CREATE INDEX idx_crowd_status_mess_id_updated_at_desc ON crowd_status (mess_id, updated_at DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_crowd_status_mess_id_updated_at_desc;
DROP TABLE IF EXISTS crowd_status CASCADE;

DROP TABLE IF EXISTS reviews CASCADE;

DROP TABLE IF EXISTS meal_items CASCADE;

DROP TABLE IF EXISTS meals CASCADE;

DROP INDEX IF EXISTS idx_dishes_hostel_group_lower_name;
DROP TABLE IF EXISTS dishes CASCADE;

DROP TABLE IF EXISTS default_items CASCADE;

DROP TABLE IF EXISTS meal_schedules CASCADE;

DROP TABLE IF EXISTS otp_codes CASCADE;

DROP TABLE IF EXISTS manager_messes CASCADE;

DROP TRIGGER IF EXISTS trigger_users_email_lowercase ON users;
DROP FUNCTION IF EXISTS set_users_email_lowercase();
DROP TABLE IF EXISTS users CASCADE;

DROP TABLE IF EXISTS messes CASCADE;
