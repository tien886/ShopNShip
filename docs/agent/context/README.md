# Implementation Context

This directory is used by the AI agent to store and retrieve context between implementation turns.

## Usage

After each significant task or at the end of a session, the AI agent should update a file here (e.g., `current_state.md`) with:
- **Summary of Changes**: What was implemented in the last turn.
- **Pending Tasks**: What needs to be done next.
- **Technical Context**: Decisions made, new dependencies, or architectural shifts.
- **References**: Specific files or lines to focus on in the next turn.

This ensures a seamless transition and prevents the AI from losing track of the implementation plan.
