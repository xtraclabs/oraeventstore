create table events (
    id  number generated always as identity,
    event_time timestamp DEFAULT current_timestamp,
    aggregate_id varchar2(60)not null,
    version integer not null,
    typecode varchar2(30) not null,
    payload blob,
    primary key(aggregate_id,version)
);

create table publish (
    aggregate_id varchar2(60)not null,
    version integer not null,
    primary key(aggregate_id,version)
);

create or replace synonym esusr.events for esdbo.events;
grant select, insert on events to esusr;

create or replace synonym esusr.publish for esdbo.publish;
grant select, insert, delete on publish to esusr;