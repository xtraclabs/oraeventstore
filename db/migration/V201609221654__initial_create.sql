create table t_aeev_events (
    id  number generated always as identity,
    event_time timestamp DEFAULT current_timestamp,
    aggregate_id varchar2(60)not null,
    version integer not null,
    typecode varchar2(30) not null,
    payload blob,
    primary key(aggregate_id,version)
);

create table t_aepb_publish (
    aggregate_id varchar2(60)not null,
    version integer not null,
    primary key(aggregate_id,version)
);