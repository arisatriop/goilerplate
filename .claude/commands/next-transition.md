# Jira Transition — Move Ticket to Next Status

Move a Jira ticket to a new status by executing an available transition.

Usage:
- `/nex-transition TICKET_ID` — pick a transition for that ticket
- `/nex-transition` — fetch tickets assigned to you, then pick one

1. Read credentials from `config/.env` using grep:
   ```bash
   grep -E "^JIRA_URL=" config/.env | cut -d= -f2-
   grep -E "^JIRA_EMAIL=" config/.env | cut -d= -f2-
   grep -E "^JIRA_API_TOKEN=" config/.env | cut -d= -f2-
   ```
   If any of the three are missing or empty, stop and tell the user which vars to add to `config/.env`.

2. If no TICKET_ID was given:
   - Fetch tickets assigned to the current user:
     ```bash
     curl -s -G "$JIRA_URL/rest/api/3/search/jql" \
       -u "$JIRA_EMAIL:$JIRA_API_TOKEN" \
       --data-urlencode "jql=assignee=currentUser() ORDER BY updated DESC" \
       --data-urlencode "fields=summary,status,priority" \
       --data-urlencode "maxResults=20"
     ```
   - Display the results as a numbered list in this format:
     ```
     #  TICKET-ID  [Status]       Priority — Summary
     1. PROJ-42    [To Do]        Medium   — Add user profile endpoint
     2. PROJ-38    [In Progress]  High     — Fix auth token expiry
     ...
     ```
   - Ask the user: "Which ticket do you want to transition? (enter number or ticket ID)". Wait for their answer, then set TICKET_ID accordingly.

3. Fetch the current issue status and available transitions:
   ```bash
   # Get current status
   curl -s "$JIRA_URL/rest/api/3/issue/$TICKET_ID?fields=summary,status" \
     -u "$JIRA_EMAIL:$JIRA_API_TOKEN"

   # Get available transitions
   curl -s "$JIRA_URL/rest/api/3/issue/$TICKET_ID/transitions" \
     -u "$JIRA_EMAIL:$JIRA_API_TOKEN"
   ```

4. Display the ticket's current status and available transitions as a numbered list:
   ```
   Ticket:  PROJ-42 — Add user profile endpoint
   Current: In Progress

   Available transitions:
   1. In Review
   2. Done
   3. Back to To Do
   ```
   Ask the user: "Which status do you want to move this ticket to? (enter number or transition name)". Wait for their answer.

5. Execute the chosen transition:
   ```bash
   curl -s -o /tmp/jira_response.json -w "%{http_code}" \
     -X POST "$JIRA_URL/rest/api/3/issue/$TICKET_ID/transitions" \
     -u "$JIRA_EMAIL:$JIRA_API_TOKEN" \
     -H "Content-Type: application/json" \
     -d "{\"transition\": {\"id\": \"$TRANSITION_ID\"}}"
   ```

6. Check the HTTP status code:
   - `204` — success. Confirm to the user: "Ticket $TICKET_ID moved to [new status]."
   - `401` — credentials invalid. Tell the user to check `JIRA_EMAIL` and `JIRA_API_TOKEN` in `config/.env`.
   - `404` — ticket not found or transition unavailable. Tell the user to verify the ticket ID.
   - Other — show the response body from `/tmp/jira_response.json` for debugging.

7. Clean up: `rm -f /tmp/jira_response.json`
