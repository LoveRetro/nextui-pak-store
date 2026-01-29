-- name: ListInstalledPaks :many
SELECT *
FROM installed_paks
WHERE can_uninstall = 1
ORDER BY name;

-- name: ListInstalledPaksWithoutRepo :many
SELECT *
FROM installed_paks
WHERE repo_url IS NULL;

-- name: ListInstalledPaksWithoutPakID :many
SELECT *
FROM installed_paks
WHERE pak_id IS NULL OR pak_id = '';

-- name: ListInstalledPaksWithPakID :many
SELECT *
FROM installed_paks
WHERE pak_id IS NOT NULL AND pak_id != '';

-- name: SyncInstalledByPakID :exec
UPDATE installed_paks
SET display_name = @display_name,
    name         = @name,
    repo_url     = @repo_url
WHERE pak_id = @pak_id;

-- name: Install :exec
INSERT INTO installed_paks (display_name, name, pak_id, repo_url, version, type, can_uninstall)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateVersion :exec
UPDATE installed_paks
SET version = ?, repo_url = ?
WHERE pak_id = ?;

-- name: SyncPakStore :exec
UPDATE installed_paks
SET display_name = @display_name,
    name         = @name,
    version      = @version,
    repo_url     = @repo_url
WHERE pak_id = @pak_id;

-- name: SyncPakStoreByName :exec
UPDATE installed_paks
SET display_name = @display_name,
    name         = @name,
    version      = @version,
    repo_url     = @repo_url,
    pak_id       = @pak_id
WHERE name = @old_name AND (pak_id IS NULL OR pak_id = '');

-- name: Uninstall :exec
DELETE
FROM installed_paks
WHERE pak_id = ? AND pak_id IS NOT NULL AND pak_id != '';

-- name: UpdateInstalledWithRepo :exec
UPDATE installed_paks
SET display_name = @new_display_name,
    name         = @new_name,
    repo_url     = @new_repo_url
WHERE display_name = @old_display_name;

-- name: UpdateInstalledWithPakID :exec
UPDATE installed_paks
SET pak_id       = @pak_id,
    display_name = @new_display_name,
    name         = @new_name,
    repo_url     = @new_repo_url
WHERE repo_url = @old_repo_url;
