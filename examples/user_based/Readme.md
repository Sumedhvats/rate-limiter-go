Each user gets their own rate limit bucket
Authentication: X-User-ID header (simplified for demo)
Real app: extract from JWT, session, etc.


# User 1: Make 21 requests (1 should fail)
for i in {1..21}; do
  curl -H "X-User-ID: alice" <http://localhost:8080/api/profile>
done

# User 2: Should have fresh limit
curl -H "X-User-ID: bob" <http://localhost:8080/api/profile>