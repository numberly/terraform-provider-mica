"""
FlashBlade S3 Bucket — Dual-Array Bidirectional Replication (Python)

Mirrors the terraform-flashblade-s3-bucket module:
  - Object store accounts on both arrays
  - S3 export policies + account exports
  - IAM-style access policies (global + per-user)
  - Named S3 users with per-user policies and access keys
  - Versioned buckets with quotas
  - Bidirectional replication via remote credentials + replica links
  - Lifecycle rules, audit filters, and QoS policy (optional)

Prerequisites:
  - Array connection configured between both arrays
  - Servers pre-provisioned on both arrays
"""

import pulumi
import pulumi_flashblade as flashblade

# ---------------------------------------------------------------------------
# Configuration (set via `pulumi config` or Pulumi.{stack}.yaml)
# ---------------------------------------------------------------------------
config = pulumi.Config()

# Array endpoints and credentials
par5_endpoint = config.require("par5_endpoint")
pa7_endpoint = config.require("pa7_endpoint")
par5_token = config.require_secret("par5_api_token")
pa7_token = config.require_secret("pa7_api_token")

# Array metadata
par5_array_name = config.require("par5_array_name")
pa7_array_name = config.require("pa7_array_name")
par5_server_name = config.require("par5_server_name")
pa7_server_name = config.require("pa7_server_name")

# Bucket settings
account_name = config.require("account_name")
bucket_name = config.require("bucket_name")
bucket_quota_bytes = config.get_int("bucket_quota_bytes")
fqdn = config.get("fqdn") or ""

# Optional features
enable_s3_target_replication = config.get_bool("enable_s3_target_replication") or False
flashblade_target_par5_sees_pa7 = config.get("flashblade_target_par5_sees_pa7") or ""
flashblade_target_pa7_sees_par5 = config.get("flashblade_target_pa7_sees_par5") or ""
audit_enabled = config.get_bool("audit_enabled") or False
audit_prefixes = config.get_object("audit_prefixes") or []

# Per-user S3 credentials (map of username -> settings)
users = config.get_object("users") or {}

# Lifecycle rules (map of rule_id -> settings)
lifecycle_rules = config.get_object("lifecycle_rules") or {
    "default": {
        "prefix": "",
        "enabled": True,
        "keep_previous_version_for": 2592000000,
        "keep_current_version_for": 2592000000,
        "abort_incomplete_multipart_uploads_after": 604800000,
    }
}

# QoS policy (null disables)
qos = config.get_object("qos")

# ---------------------------------------------------------------------------
# Providers (dual-array)
# ---------------------------------------------------------------------------
provider_par5 = flashblade.Provider("par5",
    endpoint=par5_endpoint,
    auth={"api_token": par5_token},
)

provider_pa7 = flashblade.Provider("pa7",
    endpoint=pa7_endpoint,
    auth={"api_token": pa7_token},
)

# ---------------------------------------------------------------------------
# Step 1: Verify array connections (data sources)
# ---------------------------------------------------------------------------
par5_sees_pa7 = flashblade.get_array_connection(
    remote_name=pa7_array_name,
    opts=pulumi.InvokeOptions(provider=provider_par5),
)

pa7_sees_par5 = flashblade.get_array_connection(
    remote_name=par5_array_name,
    opts=pulumi.InvokeOptions(provider=provider_pa7),
)

# ---------------------------------------------------------------------------
# Step 2: Reference existing servers
# ---------------------------------------------------------------------------
par5_server = flashblade.get_server(
    name=par5_server_name,
    opts=pulumi.InvokeOptions(provider=provider_par5),
)

pa7_server = flashblade.get_server(
    name=pa7_server_name,
    opts=pulumi.InvokeOptions(provider=provider_pa7),
)

# ---------------------------------------------------------------------------
# Step 3: Object store accounts
# ---------------------------------------------------------------------------
account_par5 = flashblade.ObjectStoreAccount("par5",
    name=account_name,
    skip_default_export=True,
    opts=pulumi.ResourceOptions(provider=provider_par5),
)

account_pa7 = flashblade.ObjectStoreAccount("pa7",
    name=account_name,
    skip_default_export=True,
    opts=pulumi.ResourceOptions(provider=provider_pa7),
)

# ---------------------------------------------------------------------------
# Step 4: S3 export policies + rules
# ---------------------------------------------------------------------------
s3_policy_par5 = flashblade.S3ExportPolicy("par5",
    name=f"{account_name}-s3-export",
    enabled=True,
    opts=pulumi.ResourceOptions(provider=provider_par5),
)

s3_policy_rule_par5 = flashblade.S3ExportPolicyRule("par5",
    policy_name=s3_policy_par5.name,
    name="allows3",
    actions=["pure:S3Access"],
    effect="allow",
    resources=["*"],
    opts=pulumi.ResourceOptions(provider=provider_par5),
)

s3_policy_pa7 = flashblade.S3ExportPolicy("pa7",
    name=f"{account_name}-s3-export",
    enabled=True,
    opts=pulumi.ResourceOptions(provider=provider_pa7),
)

s3_policy_rule_pa7 = flashblade.S3ExportPolicyRule("pa7",
    policy_name=s3_policy_pa7.name,
    name="allows3",
    actions=["pure:S3Access"],
    effect="allow",
    resources=["*"],
    opts=pulumi.ResourceOptions(provider=provider_pa7),
)

# ---------------------------------------------------------------------------
# Step 5: Account exports
# ---------------------------------------------------------------------------
account_export_par5 = flashblade.ObjectStoreAccountExport("par5",
    account_name=account_par5.name,
    server_name=par5_server.name,
    policy_name=s3_policy_par5.name,
    enabled=True,
    opts=pulumi.ResourceOptions(provider=provider_par5),
)

account_export_pa7 = flashblade.ObjectStoreAccountExport("pa7",
    account_name=account_pa7.name,
    server_name=pa7_server.name,
    policy_name=s3_policy_pa7.name,
    enabled=True,
    opts=pulumi.ResourceOptions(provider=provider_pa7),
)

# ---------------------------------------------------------------------------
# Step 6: IAM-style access policies (global rw)
# ---------------------------------------------------------------------------
access_policy_par5 = flashblade.ObjectStoreAccessPolicy("par5",
    name=f"{account_name}/rw",
    opts=pulumi.ResourceOptions(provider=provider_par5),
)

access_policy_rule_par5 = flashblade.ObjectStoreAccessPolicyRule("par5",
    policy_name=access_policy_par5.name,
    name="allowrw",
    effect="allow",
    actions=["s3:*"],
    resources=["*"],
    opts=pulumi.ResourceOptions(provider=provider_par5),
)

access_policy_pa7 = flashblade.ObjectStoreAccessPolicy("pa7",
    name=f"{account_name}/rw",
    opts=pulumi.ResourceOptions(provider=provider_pa7),
)

access_policy_rule_pa7 = flashblade.ObjectStoreAccessPolicyRule("pa7",
    policy_name=access_policy_pa7.name,
    name="allowrw",
    effect="allow",
    actions=["s3:*"],
    resources=["*"],
    opts=pulumi.ResourceOptions(provider=provider_pa7),
)

# ---------------------------------------------------------------------------
# Step 7: Buckets with versioning
# ---------------------------------------------------------------------------
bucket_par5 = flashblade.Bucket("par5",
    name=bucket_name,
    account=account_par5.name,
    versioning="enabled",
    quota_limit=bucket_quota_bytes,
    hard_limit_enabled=(bucket_quota_bytes is not None),
    destroy_eradicate_on_delete=False,
    opts=pulumi.ResourceOptions(
        provider=provider_par5,
        custom_timeouts=pulumi.CustomTimeouts(create="20m", update="20m", delete="30m"),
    ),
)

bucket_pa7 = flashblade.Bucket("pa7",
    name=bucket_name,
    account=account_pa7.name,
    versioning="enabled",
    quota_limit=bucket_quota_bytes,
    hard_limit_enabled=(bucket_quota_bytes is not None),
    destroy_eradicate_on_delete=False,
    opts=pulumi.ResourceOptions(
        provider=provider_pa7,
        custom_timeouts=pulumi.CustomTimeouts(create="20m", update="20m", delete="30m"),
    ),
)

# ---------------------------------------------------------------------------
# Step 8: Replication user, policy, and access keys
# ---------------------------------------------------------------------------
replication_user_par5 = flashblade.ObjectStoreUser("replication_par5",
    name=f"{account_name}/replication",
    full_access=True,
    opts=pulumi.ResourceOptions(provider=provider_par5),
)

replication_user_pa7 = flashblade.ObjectStoreUser("replication_pa7",
    name=f"{account_name}/replication",
    full_access=True,
    opts=pulumi.ResourceOptions(provider=provider_pa7),
)

replication_policy_par5 = flashblade.ObjectStoreAccessPolicy("replication_par5",
    name=f"{account_name}/replication",
    opts=pulumi.ResourceOptions(provider=provider_par5),
)

replication_policy_rule_par5 = flashblade.ObjectStoreAccessPolicyRule("replication_par5",
    policy_name=replication_policy_par5.name,
    name="replicationrw",
    effect="allow",
    actions=["s3:*"],
    resources=["*"],
    opts=pulumi.ResourceOptions(provider=provider_par5),
)

replication_policy_pa7 = flashblade.ObjectStoreAccessPolicy("replication_pa7",
    name=f"{account_name}/replication",
    opts=pulumi.ResourceOptions(provider=provider_pa7),
)

replication_policy_rule_pa7 = flashblade.ObjectStoreAccessPolicyRule("replication_pa7",
    policy_name=replication_policy_pa7.name,
    name="replicationrw",
    effect="allow",
    actions=["s3:*"],
    resources=["*"],
    opts=pulumi.ResourceOptions(provider=provider_pa7),
)

replication_user_policy_par5 = flashblade.ObjectStoreUserPolicy("replication_par5",
    user_name=replication_user_par5.name,
    policy_name=replication_policy_par5.name,
    opts=pulumi.ResourceOptions(provider=provider_par5),
)

replication_user_policy_pa7 = flashblade.ObjectStoreUserPolicy("replication_pa7",
    user_name=replication_user_pa7.name,
    policy_name=replication_policy_pa7.name,
    opts=pulumi.ResourceOptions(provider=provider_pa7),
)

replication_key_par5 = flashblade.ObjectStoreAccessKey("par5",
    object_store_account=account_par5.name,
    user=replication_user_par5.name,
    opts=pulumi.ResourceOptions(provider=provider_par5),
)

replication_key_pa7 = flashblade.ObjectStoreAccessKey("pa7",
    object_store_account=account_pa7.name,
    user=replication_user_pa7.name,
    name=replication_key_par5.name,
    secret_access_key=replication_key_par5.secret_access_key,
    opts=pulumi.ResourceOptions(provider=provider_pa7),
)

# ---------------------------------------------------------------------------
# Step 9: Named S3 users with per-user policies
# ---------------------------------------------------------------------------
user_resources_par5 = {}
user_resources_pa7 = {}
user_keys_par5 = {}

for username, user_cfg in users.items():
    safe_name = username.replace("-", "_")

    # Users
    user_par5 = flashblade.ObjectStoreUser(f"user_par5_{safe_name}",
        name=f"{account_name}/{username}",
        full_access=user_cfg.get("full_access", False),
        opts=pulumi.ResourceOptions(provider=provider_par5),
    )
    user_pa7 = flashblade.ObjectStoreUser(f"user_pa7_{safe_name}",
        name=f"{account_name}/{username}",
        full_access=user_cfg.get("full_access", False),
        opts=pulumi.ResourceOptions(provider=provider_pa7),
    )
    user_resources_par5[username] = user_par5
    user_resources_pa7[username] = user_pa7

    # Per-user access policies
    user_policy_par5 = flashblade.ObjectStoreAccessPolicy(f"user_par5_{safe_name}",
        name=f"{account_name}/{username}",
        opts=pulumi.ResourceOptions(provider=provider_par5),
    )
    user_policy_rule_par5 = flashblade.ObjectStoreAccessPolicyRule(f"user_par5_{safe_name}",
        policy_name=user_policy_par5.name,
        name=f"{username.replace('-', '')}rule",
        effect=user_cfg.get("effect", "allow"),
        actions=user_cfg.get("s3_actions", ["s3:GetObject", "s3:PutObject"]),
        resources=user_cfg.get("s3_resources", ["*"]),
        opts=pulumi.ResourceOptions(provider=provider_par5),
    )

    user_policy_pa7 = flashblade.ObjectStoreAccessPolicy(f"user_pa7_{safe_name}",
        name=f"{account_name}/{username}",
        opts=pulumi.ResourceOptions(provider=provider_pa7),
    )
    user_policy_rule_pa7 = flashblade.ObjectStoreAccessPolicyRule(f"user_pa7_{safe_name}",
        policy_name=user_policy_pa7.name,
        name=f"{username.replace('-', '')}rule",
        effect=user_cfg.get("effect", "allow"),
        actions=user_cfg.get("s3_actions", ["s3:GetObject", "s3:PutObject"]),
        resources=user_cfg.get("s3_resources", ["*"]),
        opts=pulumi.ResourceOptions(provider=provider_pa7),
    )

    # User-to-policy associations
    flashblade.ObjectStoreUserPolicy(f"user_par5_{safe_name}",
        user_name=user_par5.name,
        policy_name=user_policy_par5.name,
        opts=pulumi.ResourceOptions(provider=provider_par5),
    )
    flashblade.ObjectStoreUserPolicy(f"user_pa7_{safe_name}",
        user_name=user_pa7.name,
        policy_name=user_policy_pa7.name,
        opts=pulumi.ResourceOptions(provider=provider_pa7),
    )

    # Per-user access keys
    key_par5 = flashblade.ObjectStoreAccessKey(f"user_par5_{safe_name}",
        object_store_account=account_par5.name,
        user=user_par5.name,
        opts=pulumi.ResourceOptions(provider=provider_par5),
    )
    key_pa7 = flashblade.ObjectStoreAccessKey(f"user_pa7_{safe_name}",
        object_store_account=account_pa7.name,
        user=user_pa7.name,
        name=key_par5.name,
        secret_access_key=key_par5.secret_access_key,
        opts=pulumi.ResourceOptions(provider=provider_pa7),
    )
    user_keys_par5[username] = key_par5

    # Vault secret (optional — uncomment if pulumi-vault is installed)
    # import pulumi_vault as vault
    # vault_path = user_cfg.get("vault_secret_path", "")
    # if vault_path:
    #     parts = vault_path.split("/")
    #     mount = parts[0]
    #     path_rest = "/".join(parts[1:]) + f"/{username}"
    #     vault.kv.SecretV2(f"user_{safe_name}",
    #         mount=mount,
    #         name=path_rest,
    #         data_json=pulumi.Output.json.dumps({
    #             "username": username,
    #             "fqdn": fqdn,
    #             "bucket": bucket_name,
    #             "account": account_name,
    #             "access_key_id": key_par5.access_key_id,
    #             "secret_access_key": key_par5.secret_access_key,
    #         }),
    #     )

# ---------------------------------------------------------------------------
# Step 10: Remote credentials
# ---------------------------------------------------------------------------
remote_creds_par5_to_pa7 = flashblade.ObjectStoreRemoteCredentials("par5_to_pa7",
    name=f"{pa7_array_name}/{bucket_name}-creds",
    access_key_id=replication_key_pa7.access_key_id,
    secret_access_key=replication_key_pa7.secret_access_key,
    remote_name=pa7_array_name,
    opts=pulumi.ResourceOptions(provider=provider_par5),
)

remote_creds_pa7_to_par5 = flashblade.ObjectStoreRemoteCredentials("pa7_to_par5",
    name=f"{par5_array_name}/{bucket_name}-creds",
    access_key_id=replication_key_par5.access_key_id,
    secret_access_key=replication_key_par5.secret_access_key,
    remote_name=par5_array_name,
    opts=pulumi.ResourceOptions(provider=provider_pa7),
)

# ---------------------------------------------------------------------------
# Step 11: Bidirectional bucket replica links
# ---------------------------------------------------------------------------
replica_link_par5_to_pa7 = flashblade.BucketReplicaLink("par5_to_pa7",
    local_bucket_name=bucket_par5.name,
    remote_bucket_name=bucket_pa7.name,
    remote_credentials_name=remote_creds_par5_to_pa7.name,
    opts=pulumi.ResourceOptions(provider=provider_par5),
)

replica_link_pa7_to_par5 = flashblade.BucketReplicaLink("pa7_to_par5",
    local_bucket_name=bucket_pa7.name,
    remote_bucket_name=bucket_par5.name,
    remote_credentials_name=remote_creds_pa7_to_par5.name,
    opts=pulumi.ResourceOptions(provider=provider_pa7),
)

# ---------------------------------------------------------------------------
# Step 11b: Optional S3 target replication
# ---------------------------------------------------------------------------
if enable_s3_target_replication:
    if not flashblade_target_par5_sees_pa7 or not flashblade_target_pa7_sees_par5:
        raise ValueError(
            "flashblade_target_par5_sees_pa7 and flashblade_target_pa7_sees_par5 "
            "must be set when enable_s3_target_replication is true"
        )
    remote_creds_par5_to_pa7_s3 = flashblade.ObjectStoreRemoteCredentials("par5_to_pa7_s3",
        name=f"{flashblade_target_par5_sees_pa7}/{bucket_name}-creds",
        access_key_id=replication_key_pa7.access_key_id,
        secret_access_key=replication_key_pa7.secret_access_key,
        target_name=flashblade_target_par5_sees_pa7,
        opts=pulumi.ResourceOptions(provider=provider_par5),
    )
    remote_creds_pa7_to_par5_s3 = flashblade.ObjectStoreRemoteCredentials("pa7_to_par5_s3",
        name=f"{flashblade_target_pa7_sees_par5}/{bucket_name}-creds",
        access_key_id=replication_key_par5.access_key_id,
        secret_access_key=replication_key_par5.secret_access_key,
        target_name=flashblade_target_pa7_sees_par5,
        opts=pulumi.ResourceOptions(provider=provider_pa7),
    )

    flashblade.BucketReplicaLink("par5_to_pa7_s3",
        local_bucket_name=bucket_par5.name,
        remote_bucket_name=bucket_pa7.name,
        remote_credentials_name=remote_creds_par5_to_pa7_s3.name,
        opts=pulumi.ResourceOptions(provider=provider_par5),
    )
    flashblade.BucketReplicaLink("pa7_to_par5_s3",
        local_bucket_name=bucket_pa7.name,
        remote_bucket_name=bucket_par5.name,
        remote_credentials_name=remote_creds_pa7_to_par5_s3.name,
        opts=pulumi.ResourceOptions(provider=provider_pa7),
    )

# ---------------------------------------------------------------------------
# Step 12: Lifecycle rules
# ---------------------------------------------------------------------------
for rule_id, rule_cfg in lifecycle_rules.items():
    safe_rule = rule_id.replace("-", "_")
    flashblade.LifecycleRule(f"par5_{safe_rule}",
        bucket_name=bucket_par5.name,
        rule_id=rule_id,
        prefix=rule_cfg.get("prefix", ""),
        enabled=rule_cfg.get("enabled", True),
        keep_previous_version_for=rule_cfg.get("keep_previous_version_for"),
        keep_current_version_for=rule_cfg.get("keep_current_version_for"),
        keep_current_version_until=rule_cfg.get("keep_current_version_until"),
        abort_incomplete_multipart_uploads_after=rule_cfg.get("abort_incomplete_multipart_uploads_after"),
        opts=pulumi.ResourceOptions(provider=provider_par5),
    )
    flashblade.LifecycleRule(f"pa7_{safe_rule}",
        bucket_name=bucket_pa7.name,
        rule_id=rule_id,
        prefix=rule_cfg.get("prefix", ""),
        enabled=rule_cfg.get("enabled", True),
        keep_previous_version_for=rule_cfg.get("keep_previous_version_for"),
        keep_current_version_for=rule_cfg.get("keep_current_version_for"),
        keep_current_version_until=rule_cfg.get("keep_current_version_until"),
        abort_incomplete_multipart_uploads_after=rule_cfg.get("abort_incomplete_multipart_uploads_after"),
        opts=pulumi.ResourceOptions(provider=provider_pa7),
    )

# ---------------------------------------------------------------------------
# Step 13: Audit filters (optional)
# ---------------------------------------------------------------------------
if audit_enabled:
    flashblade.BucketAuditFilter("par5",
        name="auditwrite",
        bucket_name=bucket_par5.name,
        actions=["s3:PutObject", "s3:DeleteObject"],
        s3_prefixes=audit_prefixes if audit_prefixes else [],
        opts=pulumi.ResourceOptions(provider=provider_par5),
    )
    flashblade.BucketAuditFilter("pa7",
        name="auditwrite",
        bucket_name=bucket_pa7.name,
        actions=["s3:PutObject", "s3:DeleteObject"],
        s3_prefixes=audit_prefixes if audit_prefixes else [],
        opts=pulumi.ResourceOptions(provider=provider_pa7),
    )

# ---------------------------------------------------------------------------
# Step 14: QoS policy (optional)
# ---------------------------------------------------------------------------
if qos:
    flashblade.QosPolicy("this",
        name=f"{account_name}-qos",
        enabled=True,
        max_total_bytes_per_sec=qos.get("max_total_bytes_per_sec"),
        max_total_ops_per_sec=qos.get("max_total_ops_per_sec"),
        opts=pulumi.ResourceOptions(provider=provider_par5),
    )

# ---------------------------------------------------------------------------
# Outputs
# ---------------------------------------------------------------------------
pulumi.export("par5_connection_status", par5_sees_pa7.status)
pulumi.export("pa7_connection_status", pa7_sees_par5.status)
pulumi.export("par5_bucket_id", bucket_par5.id)
pulumi.export("pa7_bucket_id", bucket_pa7.id)
pulumi.export("par5_replica_status", replica_link_par5_to_pa7.status)
pulumi.export("pa7_replica_status", replica_link_pa7_to_par5.status)
pulumi.export("fqdn", fqdn)

# Per-user access key IDs
pulumi.export("user_access_key_ids", {
    username: key.access_key_id
    for username, key in user_keys_par5.items()
})
