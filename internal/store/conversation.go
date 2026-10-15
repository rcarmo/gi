package store

// These SQL expressions apply only to the opt-in conversation view. Raw message
// storage/export and model history keep their original content and roles.
const conversationContentSQL = `case when role='assistant' and json_extract(payload_json,'$.kind')='tool_calls' then
 case when json_type(payload_json,'$.display_text')='text' then json_extract(payload_json,'$.display_text')
 when substr(content,1,11)='[tool_call:' then ''
 when instr(content,char(10)||'[tool_call:')>0 then substr(content,1,instr(content,char(10)||'[tool_call:')-1)
 else content end else content end`

const conversationVisibleSQL = `(role in ('user','assistant','system') or (role='tool_result' and json_extract(payload_json,'$.kind')='tool_result' and json_extract(payload_json,'$.tool_name')='codemode'))
 and (role='user' or coalesce(json_extract(payload_json,'$.kind'),'') != 'tool_result' or (role='tool_result' and json_extract(payload_json,'$.tool_name')='codemode'))
 and not (role='system' and json_extract(payload_json,'$.kind') is 'queue')
 and not (role='assistant' and json_extract(payload_json,'$.kind') is 'tool_calls'
 and trim(` + conversationContentSQL + `)='')`

// conversationReplyToSQL projects an assistant reply onto the first user
// message in its native turn. It does not change stored/model message data.
// Session scoping is necessary because caller-provided turn IDs are not global.
func conversationReplyToSQL(alias string) string {
	return `case when ` + alias + `.role='assistant' and
 json_type(` + alias + `.payload_json,'$.turn_id')='text' and
 json_extract(` + alias + `.payload_json,'$.turn_id')!='' then coalesce((
 select parent.id from messages parent where parent.session_id=` + alias + `.session_id
 and parent.role='user' and json_extract(parent.payload_json,'$.turn_id')=json_extract(` + alias + `.payload_json,'$.turn_id')
 order by parent.created_at,parent.id limit 1),'') else '' end`
}

// Remove replay blobs before driver decoding, so large image payloads are not
// allocated by display-only reads. Raw/export queries still select payload_json.
const conversationPayloadSQL = `json_remove(payload_json,'$.provider_message','$.provider_content_blocks','$.added_tool_names')`

// conversationPayload removes internal replay envelopes from UI projections.
// Raw exports and model snapshots retain the complete stored payload.
func conversationPayload(payload map[string]any) map[string]any {
	out := make(map[string]any, len(payload))
	for key, value := range payload {
		switch key {
		case "provider_message", "provider_content_blocks", "added_tool_names":
			continue
		}
		out[key] = value
	}
	return out
}
