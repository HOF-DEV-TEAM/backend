# /jira — Jira Ticket Workflow

Manage Jira tickets through the full development lifecycle.
Arguments: $ARGUMENTS — ticket ID (e.g. HOF-123), action (e.g. "start", "review", "done"), or "new".

---

## Fetch a ticket

```bash
# View ticket details
jira issue view HOF-$TICKET_ID

# Open in browser
jira issue open HOF-$TICKET_ID

# List tickets assigned to me
jira issue list --assignee $(jira me) --status "In Progress"

# List all open tickets in the current sprint
jira sprint list --current
jira issue list --sprint "$(jira sprint list --current -q --plain | head -1)"
```

---

## Ticket lifecycle

### Start working

```bash
# Move to In Progress
jira issue transition HOF-$TICKET_ID "In Progress"

# Assign to yourself
jira issue assign HOF-$TICKET_ID $(jira me)

# Create the feature branch
git checkout -b hof-$TICKET_ID-short-description master
# Example: git checkout -b hof-123-restore-webhook-events master
```

Branch naming convention:
```
hof-<ticket-id>-<short-description-kebab-case>
```

### During development

```bash
# Add a comment with progress update
jira issue comment HOF-$TICKET_ID "Work in progress — implementing webhook event handlers"

# Log work (if time tracking enabled)
jira issue worklog add HOF-$TICKET_ID --time 2h
```

### Ready for review

```bash
# Move to Code Review
jira issue transition HOF-$TICKET_ID "In Review"

# Comment with PR link
PR_URL=$(gh pr view --json url -q .url)
jira issue comment HOF-$TICKET_ID "PR opened for review: $PR_URL"

# Link the PR to the ticket
jira issue link HOF-$TICKET_ID "GitHub PR" "$PR_URL"
```

### After merge

```bash
# Move to Done
jira issue transition HOF-$TICKET_ID "Done"

# Or to QA if there's a QA stage
jira issue transition HOF-$TICKET_ID "QA"
```

---

## Creating a new ticket

```bash
jira issue create \
  --project HOF \
  --type "Story" \
  --summary "Add phone number to user profile" \
  --description "$(cat <<'EOF'
As a user, I want to add my phone number to my profile so that I can receive SMS notifications.

**Acceptance Criteria:**
- [ ] Users can set/update phone number via PATCH /user/update
- [ ] Phone is returned in user profile response
- [ ] Phone is optional (nullable)
- [ ] Migration adds phone column to users table

**Technical Notes:**
- Add migration: 026_add_phone_to_users.sql
- Update User entity: Phone *string
- Update UpdateProfileRequest DTO
EOF
)"
```

---

## Linking tickets to commits

Include the ticket ID in every commit message:
```
feat(user): add phone number to profile [HOF-124]

Adds optional phone field to User entity, UpdateProfileRequest,
and the users table via migration 026.
```

When using `/commit`, always include `[HOF-<ID>]` at the end of the subject line.

---

## Ticket states for HOF

| State | Meaning |
|---|---|
| Backlog | Not yet started |
| To Do | Prioritised for current sprint |
| In Progress | Actively being worked on |
| In Review | PR open, awaiting code review |
| QA | Merged, awaiting testing |
| Done | Shipped |
| Blocked | Waiting on dependency or decision |

---

## Jira CLI setup (if not configured)

```bash
# Install
brew install ankitpokhrel/jira-cli/jira-cli   # macOS
# or: go install github.com/ankitpokhrel/jira-cli/cmd/jira@latest

# Configure
jira init

# Provide:
# - Jira server URL: https://<your-org>.atlassian.net
# - Auth: API token (generate at id.atlassian.com/manage-profile/security/api-tokens)
# - Project: HOF
```
