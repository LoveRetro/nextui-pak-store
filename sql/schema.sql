create table installed_paks
(
    name          text not null,
    display_name  text not null,
    pak_id        text,
    repo_url      text,
    type          text not null,
    version       text not null,
    can_uninstall int  not null,
    unique (name)
);