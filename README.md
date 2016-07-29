Library setup:



<pre>
export PKG_CONFIG_PATH=$GOPATH/src/github.com/xtraclabs/oraeventstore/pkgconfig/
go get github.com/rjeczalik/pkgconfig/cmd/pkg-config
go get -u github.com/mattn/go-oci8
</pre>


Initial set up - login as system/oracle, and create the dbo user for the rest of the setup...

<pre>
create user esdbo
identified by password
default tablespace users
temporary tablespace temp;

grant dba to esdbo;
</pre>




create table events (
    id  number generated always as identity,
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

create user esusr
identified by password
default tablespace users
temporary tablespace temp;

create or replace synonym esusr.events for esdbo.events
grant select, insert on events to esusr;
grant connect to esusr;

create or replace synonym esusr.publish for esdbo.publish