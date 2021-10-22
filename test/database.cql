-- creates a test database with some initial data
create keyspace if not exists test WITH REPLICATION = {
    'class': 'NetworkTopologyStrategy',
    'DC1' : 3
};

use test;

create table if not exists users
(
    uuid
    timeuuid
    primary
    key,
    id text
);

create index if not exists users_id_idx
    on users (id);

INSERT INTO users (uuid, id) VALUES (447753b7-5979-11eb-ab9c-02420afc504b, '1');
INSERT INTO users (uuid, id) VALUES (defdfd20-d89a-11eb-b2ba-02420afc0259, '2');
INSERT INTO users (uuid, id) VALUES (eca00911-5998-11eb-ba88-02420afc0c6a, '3');
INSERT INTO users (uuid, id) VALUES (0ab828b6-df31-11eb-90f7-02420afc0225, '4');
INSERT INTO users (uuid, id) VALUES (94593a4c-599a-11eb-842d-02420afc0c6a, '5');