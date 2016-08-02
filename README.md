## Go Lang Set Up

Note that the settings in pkgconfig/oci8.pc need to be correct or the go get of
the go-oci8 package will fail.

<pre>
export PKG_CONFIG_PATH=$GOPATH/src/github.com/xtraclabs/oraeventstore/pkgconfig/
go get github.com/rjeczalik/pkgconfig/cmd/pkg-config
go get -u github.com/mattn/go-oci8
</pre>

## Database Set Up

Initial set up - login as system/oracle, and create the dbo user for the rest of the setup...

<pre>
create user esdbo
identified by password
default tablespace users
temporary tablespace temp;

grant dba to esdbo;
</pre>


Tables, create as esdbo:

<pre>
create table events (
    id  number generated always as identity,
    event_time timestamp DEFAULT current_timestamp,
    aggregate_id varchar2(60)not null,
    version integer not null,
    typecode varchar2(30) not null,
    payload blob,
    primary key(aggregate_id,version)
)

create table publish (
    aggregate_id varchar2(60)not null,
    version integer not null,
    primary key(aggregate_id,version)
);
</pre>

Create a user to access the tables.

<pre>
create user esusr
identified by password
default tablespace users
temporary tablespace temp;

grant connect to esusr;

create or replace synonym esusr.events for esdbo.events;
grant select, insert on events to esusr;

create or replace synonym esusr.publish for esdbo.publish;
grant select, insert on publish to esusr;
</pre>

## A Note on the Publish Table

The publish table simply writes the aggregate IDs of recently stored
aggregates, which picks up creation and updates. Another process will need
to read from the table to pick up the published aggregate, read the
actual data from the event store table, do something with it (publish it
to a queue, write out CQRS query views, etc), then delete the record from the
publish table.