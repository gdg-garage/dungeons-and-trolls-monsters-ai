/*
 * Dungeons and Trolls
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: 1.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type DungeonsandtrollsSkillEffect struct {
	Attributes *DungeonsandtrollsSkillAttributes `json:"attributes,omitempty"`
	Flags []string `json:"flags,omitempty"`
	Summons []DungeonsandtrollsDroppable `json:"summons,omitempty"`
}
