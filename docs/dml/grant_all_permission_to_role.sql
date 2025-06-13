insert into menu_permission_roles 
(role_id, menu_permission_id, created_at, created_by, updated_at, updated_by)
select 
'c37bf035-f3ea-49ec-85d6-da29bb345f3e' as role_id, 
id as menu_permission_id, 
now(), 
'system', 
now(), 
'system' 
from menu_permissions mp where mp.level is null or mp."level"  >= 1;