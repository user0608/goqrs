create table account
(
    username         varchar(80) not null,
    first_name       varchar(80) not null,
    last_name        varchar(80),
    email            varchar(80) not null unique,
    password         varchar(80) not null,
    password_attempt int         not null default 0,
    is_disabled      bool        not null default false,
    created_at       timestamp   not null default current_timestamp,
    constraint pk_account primary key (username)
);

create table collection
(
    id               uuid        not null,
    name             varchar(80) not null,
    description      varchar(480),
    time_out         timestamp,
    not_before       timestamp   not null default current_timestamp,
    created_at       timestamp   not null default current_timestamp,
    num_tickets      int         not null,
    template_uuid    varchar(80) not null default '',
    template_details json,
    document_uuid    uuid,
    document_process varchar(80) not null default '',
    process_result   varchar(80),
    account_username varchar(80) not null,
    deleted_at       timestamp,
    constraint pk_collection primary key (id),
    constraint chk_document_process check ( document_process in ('', 'processing', 'processed', 'error') ),
    constraint fk_collection__account foreign key (account_username) references account (username)
);


create table tag
(
    id            uuid         not null,
    collection_id uuid         not null,
    name          varchar(80)  not null,
    value         varchar(280) not null,
    constraint pk_tag primary key (id),
    constraint fk_tag__collection foreign key (collection_id) references collection (id)
);

create table ticket
(
    id            uuid not null,
    reclaimed     timestamp,
    annulled      timestamp,
    collection_id uuid not null,
    constraint pk_ticket primary key (id),
    constraint fk_ticket__collection foreign key (collection_id) references collection (id)
);