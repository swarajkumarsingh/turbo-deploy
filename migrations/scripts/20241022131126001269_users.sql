CREATE TYPE plan_type_enum AS ENUM ('free', 'paid', 'trial');
CREATE TYPE user_role_enum AS ENUM ('admin', 'user', 'moderator');
CREATE TYPE primary_goal_enum AS ENUM ('test', 'production', 'sample');

CREATE TABLE IF NOT EXISTS users (
	id SERIAL NOT NULL PRIMARY KEY,
	username VARCHAR(500) UNIQUE NOT NULL,
	firstname VARCHAR(100) NOT NULL,
	lastname VARCHAR(100) NOT NULL,
	email VARCHAR(100) UNIQUE NOT NULL,
	password VARCHAR(200) NOT NULL,
	phone VARCHAR(12) NOT NULL,
	location TEXT,
	address TEXT,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    is_deleted BOOLEAN DEFAULT TRUE NOT NULL,
    primary_goal primary_goal_enum DEFAULT 'test' NOT NULL,
    experience INT DEFAULT 0,
    user_role user_role_enum DEFAULT 'user' NOT NULL,
    plan_type plan_type_enum DEFAULT 'free' NOT NULL,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NOW()
);
