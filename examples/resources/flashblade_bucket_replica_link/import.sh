# Import requires the replica link UUID (not bucket names), because multiple
# links can exist for the same bucket pair (e.g., array replication + S3 target).
#
# To find the UUID, query the FlashBlade API:
#
#   curl -s -k -H "x-auth-token: $FB_API_TOKEN" \
#     "https://$FB_HOST/api/2.22/bucket-replica-links?local_bucket_names=my-bucket&remote_bucket_names=my-bucket" \
#     | jq '.items[] | {id, local_bucket: .local_bucket.name, remote_bucket: .remote_bucket.name, remote_credentials: .remote_credentials.name}'
#
terraform import flashblade_bucket_replica_link.example 10000000-0000-0000-0000-000000000001
