-- Pay Plans
CREATE TABLE IF NOT EXISTS pay_plans (
	id INT GENERATED ALWAYS AS IDENTITY,
	plan_type VARCHAR NOT NULL UNIQUE,
	daily_limit INT NOT NULL,
	PRIMARY KEY (plan_type)
);

-- Blockchains
CREATE TABLE IF NOT EXISTS blockchains (
	id INT GENERATED ALWAYS AS IDENTITY,
	blockchain_id VARCHAR NOT NULL UNIQUE,
	active VARCHAR,
	altruist VARCHAR,
	blockchain VARCHAR,
	blockchain_aliases VARCHAR[],
	chain_id VARCHAR,
	chain_id_check VARCHAR,
	description VARCHAR,
	enforce_result VARCHAR,
	log_limit_blocks INT,
	network VARCHAR,
	path VARCHAR,
	request_timeout INT,
	ticker VARCHAR,
	created_at TIMESTAMP NULL,
	updated_at TIMESTAMP NULL,
	PRIMARY KEY (blockchain_id)
);

CREATE TABLE IF NOT EXISTS redirects (
	id INT GENERATED ALWAYS AS IDENTITY,
	blockchain_id VARCHAR NOT NULL,
	alias VARCHAR NOT NULL,
	loadbalancer VARCHAR NOT NULL,
	domain VARCHAR NOT NULL,
	created_at TIMESTAMP NULL,
	updated_at TIMESTAMP NULL,
	UNIQUE (blockchain_id, domain),
	PRIMARY KEY (id),
	CONSTRAINT fk_blockchain
      FOREIGN KEY(blockchain_id) 
	  	REFERENCES blockchains(blockchain_id)
);

CREATE TABLE IF NOT EXISTS sync_check_options (
	id INT GENERATED ALWAYS AS IDENTITY,
	blockchain_id VARCHAR NOT NULL UNIQUE,
	syncCheck VARCHAR,
	allowance INT,
	body VARCHAR,
	path VARCHAR,
	result_key VARCHAR,
	PRIMARY KEY (id),
	CONSTRAINT fk_blockchain
      FOREIGN KEY(blockchain_id)
	  	REFERENCES blockchains(blockchain_id)
);

-- Load Balancers
CREATE TABLE IF NOT EXISTS loadbalancers (
	id INT GENERATED ALWAYS AS IDENTITY,
	lb_id VARCHAR NOT NULL UNIQUE,
	user_id VARCHAR,
	name VARCHAR,
	request_timeout INT,
	gigastake BOOLEAN,
	gigastake_redirect BOOLEAN,
	created_at TIMESTAMP NULL,
	updated_at TIMESTAMP NULL,
	PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS stickiness_options (
	id INT GENERATED ALWAYS AS IDENTITY,
	lb_id VARCHAR NOT NULL UNIQUE,
	duration TEXT,
	sticky_max INT,
	stickiness BOOLEAN,
	origins VARCHAR[],
	PRIMARY KEY (id),
	CONSTRAINT fk_lb
      FOREIGN KEY(lb_id)
	  	REFERENCES loadbalancers(lb_id)
);

-- Applications
CREATE TABLE IF NOT EXISTS applications (
	id INT GENERATED ALWAYS AS IDENTITY,
	application_id VARCHAR NOT NULL UNIQUE,
	contact_email VARCHAR,
	description TEXT,
	name VARCHAR,
	status VARCHAR,
	pay_plan_type VARCHAR,
	owner VARCHAR,
	url VARCHAR,
	user_id VARCHAR,
	dummy BOOLEAN,
	first_date_surpassed TIMESTAMP NULL,
	created_at TIMESTAMP NULL,
	updated_at TIMESTAMP NULL,
	PRIMARY KEY (application_id),
	CONSTRAINT fk_pay_plan
      FOREIGN KEY(pay_plan_type) 
	  	REFERENCES pay_plans(plan_type)
);

CREATE TABLE IF NOT EXISTS gateway_aat (
	id INT GENERATED ALWAYS AS IDENTITY,
	application_id VARCHAR NOT NULL UNIQUE,
	address VARCHAR NOT NULL,
	public_key VARCHAR NOT NULL,
	private_key VARCHAR,
	signature VARCHAR NOT NULL,
	client_public_key VARCHAR NOT NULL,
	version VARCHAR,
	PRIMARY KEY (id),
	CONSTRAINT fk_application
      FOREIGN KEY(application_id) 
	  	REFERENCES applications(application_id)
);

CREATE TABLE IF NOT EXISTS gateway_settings (
	id INT GENERATED ALWAYS AS IDENTITY,
	application_id VARCHAR NOT NULL UNIQUE,
	secret_key VARCHAR,
	secret_key_required BOOLEAN,
	whitelist_blockchains VARCHAR[],
	whitelist_contracts VARCHAR,
	whitelist_methods VARCHAR,
	whitelist_origins VARCHAR[],
	whitelist_user_agents VARCHAR[],
	PRIMARY KEY (id),
	CONSTRAINT fk_application
      FOREIGN KEY(application_id) 
	  	REFERENCES applications(application_id)
);

CREATE TABLE IF NOT EXISTS notification_settings (
	id INT GENERATED ALWAYS AS IDENTITY,
	application_id VARCHAR NOT NULL UNIQUE,
	signed_up BOOLEAN,
	on_quarter BOOLEAN,
	on_half BOOLEAN,
	on_three_quarters BOOLEAN,
	on_full BOOLEAN,
	PRIMARY KEY (id),
	CONSTRAINT fk_application
      FOREIGN KEY(application_id) 
	  	REFERENCES applications(application_id)
);

-- Load Balancer-Apps Join Table
CREATE TABLE IF NOT EXISTS lb_apps (
	id INT GENERATED ALWAYS AS IDENTITY,
	lb_id VARCHAR NOT NULL,
	app_id VARCHAR NOT NULL,
	UNIQUE(lb_id, app_id),
	PRIMARY KEY (id),
	CONSTRAINT fk_lb
      FOREIGN KEY(lb_id) 
	  	REFERENCES loadbalancers(lb_id),
	CONSTRAINT fk_app
      FOREIGN KEY(app_id) 
	  	REFERENCES applications(application_id)
);

-- Insert Rows
INSERT INTO pay_plans (plan_type, daily_limit)
VALUES
    ('FREETIER_V0', 250000),
    ('PAY_AS_YOU_GO_V0', 0),
    ('TEST_PLAN_V0', 100),
    ('TEST_PLAN_10K', 10000),
    ('TEST_PLAN_90K', 90000);
