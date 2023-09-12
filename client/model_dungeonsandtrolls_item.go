/*
 * Dungeons and Trolls
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: 1.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type DungeonsandtrollsItem struct {
	Id string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slot *DungeonsandtrollsItemType `json:"slot,omitempty"`
	BuyPrice int32 `json:"buyPrice,omitempty"`
	Requirements *DungeonsandtrollsAttributes `json:"requirements,omitempty"`
	Attributes *DungeonsandtrollsAttributes `json:"attributes,omitempty"`
	Skills []DungeonsandtrollsSkill `json:"skills,omitempty"`
}
