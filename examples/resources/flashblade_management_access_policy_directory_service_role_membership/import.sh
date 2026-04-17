# Import composite ID format: <role_name>/<policy_name>
#
# DOC-02 supersession (D-05): REQUIREMENTS.md DOC-02 originally called for the format
# <policy_name>:<role_name>, but that legacy format is superseded by D-05. Reason:
# built-in policy names contain both ':' and '/' (e.g. pure:policy/array_admin), which
# breaks strings.SplitN(id, sep, 2) whether sep is ':' or '/' when the policy name
# comes first. Putting role_name FIRST (a simple identifier with no special chars) and
# using '/' as the separator lets SplitN("<role>/pure:policy/<pol>", "/", 2) correctly
# return ["<role>", "pure:policy/<pol>"].
#
# Canonical form: Role comes FIRST, then '/', then the full policy name (colons and
# slashes preserved verbatim).
terraform import flashblade_management_access_policy_directory_service_role_membership.example array_admin/pure:policy/array_admin
