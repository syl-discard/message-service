CREATE KEYSPACE IF NOT EXISTS messages WITH REPLICATION = {'class': 'NetworkTopologyStrategy', 'DC1': '1'};

cqlsh message-db -e "CREATE TABLE IF NOT EXISTS messages.messages (id uuid,message text, serverid uuid, userid uuid, PRIMARY KEY (id));"
