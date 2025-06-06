
TEMP_ACCESS_TOKEN=$(curl -X POST -H "Content-Type: application/json" \
                        -d '{"email": "admin@example.com", "password": "d1r3ctu5"}' \
                        $DIRECTUS_URL/auth/login \
                        | jq .data.access_token | cut -d '"' -f2)

USER_ID=$(curl -X GET -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TEMP_ACCESS_TOKEN" \
    $DIRECTUS_URL/users/me | jq .data.id | cut -d '"' -f2)

# Setting access token

curl -X PATCH -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TEMP_ACCESS_TOKEN" \
    -d "{\"token\": \"$ADMIN_ACCESS_TOKEN\"}" \
    $DIRECTUS_URL/users/$USER_ID

# ssbbot_chat_settings table
curl -X POST -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
    -d '{"collection":"ssbbot_chat_settings","fields":[{"field":"chat_id","type":"bigInteger","meta":{"hidden":true,"interface":"input","readonly":true},"schema":{"is_primary_key":true,"has_auto_increment":true}},{"field":"date_created","type":"timestamp","meta":{"special":["date-created"],"interface":"datetime","readonly":true,"hidden":true,"width":"half","display":"datetime","display_options":{"relative":true}},"schema":{}},{"field":"date_updated","type":"timestamp","meta":{"special":["date-updated"],"interface":"datetime","readonly":true,"hidden":true,"width":"half","display":"datetime","display_options":{"relative":true}},"schema":{}}],"schema":{},"meta":{"singleton":false}}' \
    $DIRECTUS_URL/collections

curl -X POST -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
    -d '{"type":"integer","meta":{"interface":"input","special":null},"field":"latest_ssb_month_notified"}' \
    $DIRECTUS_URL/fields/ssbbot_chat_settings

