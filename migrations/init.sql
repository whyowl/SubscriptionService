CREATE TABLE IF NOT EXISTS subscriptions (
                                             user_id UUID NOT NULL,
                                             service_name TEXT NOT NULL,
                                             price INTEGER NOT NULL CHECK (price >= 0),
                                             start_date DATE NOT NULL CHECK (EXTRACT(DAY FROM start_date) = 1),
                                             end_date DATE CHECK (EXTRACT(DAY FROM end_date) = 1),

                                             CONSTRAINT subscriptions_pk PRIMARY KEY (user_id, service_name)
);