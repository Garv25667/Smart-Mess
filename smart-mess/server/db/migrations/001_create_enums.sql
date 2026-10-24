-- +goose Up
-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enums
CREATE TYPE hostel_group AS ENUM ('girls', 'boys');
CREATE TYPE user_role AS ENUM ('student', 'main_manager', 'mess_manager');
CREATE TYPE meal_type AS ENUM ('breakfast', 'lunch', 'snacks', 'dinner');
CREATE TYPE day_kind AS ENUM ('weekday', 'weekend');
CREATE TYPE diet_menu AS ENUM ('regular', 'jain', 'both');
CREATE TYPE diet_pref AS ENUM ('regular', 'jain');
CREATE TYPE crowd_level AS ENUM ('low', 'medium', 'high');
CREATE TYPE meal_status AS ENUM ('draft', 'published');
CREATE TYPE cvd_mode AS ENUM ('off', 'unsure', 'protan', 'deutan', 'tritan', 'mono');
CREATE TYPE theme_pref AS ENUM ('system', 'light', 'dark', 'high_contrast');

-- +goose Down
DROP TYPE IF EXISTS theme_pref;
DROP TYPE IF EXISTS cvd_mode;
DROP TYPE IF EXISTS meal_status;
DROP TYPE IF EXISTS crowd_level;
DROP TYPE IF EXISTS diet_pref;
DROP TYPE IF EXISTS diet_menu;
DROP TYPE IF EXISTS day_kind;
DROP TYPE IF EXISTS meal_type;
DROP TYPE IF EXISTS user_role;
DROP TYPE IF EXISTS hostel_group;
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";
