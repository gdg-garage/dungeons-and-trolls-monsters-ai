# {{classname}}

All URIs are relative to *https://dt.garage-trip.cz/*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DungeonsAndTrollsBuy**](DungeonsAndTrollsApi.md#DungeonsAndTrollsBuy) | **Post** /v1/buy | Buy Items identified by the provided ID for the Character bound to the logged user.
[**DungeonsAndTrollsCommands**](DungeonsAndTrollsApi.md#DungeonsAndTrollsCommands) | **Post** /v1/commands | Send multiple commands to the Character bound to the logged user. The order of execution is defined in the message.
[**DungeonsAndTrollsGame**](DungeonsAndTrollsApi.md#DungeonsAndTrollsGame) | **Get** /v1/game | Sends all info about the game.
[**DungeonsAndTrollsMonstersCommands**](DungeonsAndTrollsApi.md#DungeonsAndTrollsMonstersCommands) | **Post** /v1/monsters-commands | Control monsters. Admin only.
[**DungeonsAndTrollsMove**](DungeonsAndTrollsApi.md#DungeonsAndTrollsMove) | **Post** /v1/move | Assign skill point to the attribute for the Character bound to the logged user.
[**DungeonsAndTrollsPickUp**](DungeonsAndTrollsApi.md#DungeonsAndTrollsPickUp) | **Post** /v1/pick-up | Equip the Item from the ground identified by the provided ID for the Character bound to the logged user (unused).
[**DungeonsAndTrollsRegister**](DungeonsAndTrollsApi.md#DungeonsAndTrollsRegister) | **Post** /v1/register | Register provided User to the Game and create a character.
[**DungeonsAndTrollsRespawn**](DungeonsAndTrollsApi.md#DungeonsAndTrollsRespawn) | **Post** /v1/respawn | Respawn the Character bound to the logged user.
[**DungeonsAndTrollsSkill**](DungeonsAndTrollsApi.md#DungeonsAndTrollsSkill) | **Post** /v1/skill | Use a skill (provided by an item) by the Character bound to the logged user.
[**DungeonsAndTrollsYell**](DungeonsAndTrollsApi.md#DungeonsAndTrollsYell) | **Post** /v1/yell | The Character bound to the logged user yells a messages (visible for everyone).

# **DungeonsAndTrollsBuy**
> interface{} DungeonsAndTrollsBuy(ctx, body)
Buy Items identified by the provided ID for the Character bound to the logged user.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**DungeonsandtrollsIdentifiers**](DungeonsandtrollsIdentifiers.md)|  | 

### Return type

[**interface{}**](interface{}.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DungeonsAndTrollsCommands**
> interface{} DungeonsAndTrollsCommands(ctx, body)
Send multiple commands to the Character bound to the logged user. The order of execution is defined in the message.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**DungeonsandtrollsCommandsBatch**](DungeonsandtrollsCommandsBatch.md)|  | 

### Return type

[**interface{}**](interface{}.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DungeonsAndTrollsGame**
> DungeonsandtrollsGameState DungeonsAndTrollsGame(ctx, optional)
Sends all info about the game.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***DungeonsAndTrollsApiDungeonsAndTrollsGameOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DungeonsAndTrollsApiDungeonsAndTrollsGameOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ascii** | **optional.Bool**| default false | 
 **events** | **optional.Bool**| default false | 
 **blocking** | **optional.Bool**| default true | 

### Return type

[**DungeonsandtrollsGameState**](dungeonsandtrollsGameState.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DungeonsAndTrollsMonstersCommands**
> interface{} DungeonsAndTrollsMonstersCommands(ctx, body)
Control monsters. Admin only.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**DungeonsandtrollsCommandsForMonsters**](DungeonsandtrollsCommandsForMonsters.md)|  | 

### Return type

[**interface{}**](interface{}.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DungeonsAndTrollsMove**
> interface{} DungeonsAndTrollsMove(ctx, body)
Assign skill point to the attribute for the Character bound to the logged user.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**DungeonsandtrollsCoordinates**](DungeonsandtrollsCoordinates.md)|  | 

### Return type

[**interface{}**](interface{}.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DungeonsAndTrollsPickUp**
> interface{} DungeonsAndTrollsPickUp(ctx, body)
Equip the Item from the ground identified by the provided ID for the Character bound to the logged user (unused).

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**DungeonsandtrollsIdentifier**](DungeonsandtrollsIdentifier.md)|  | 

### Return type

[**interface{}**](interface{}.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DungeonsAndTrollsRegister**
> DungeonsandtrollsRegistration DungeonsAndTrollsRegister(ctx, body)
Register provided User to the Game and create a character.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**DungeonsandtrollsUser**](DungeonsandtrollsUser.md)|  | 

### Return type

[**DungeonsandtrollsRegistration**](dungeonsandtrollsRegistration.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DungeonsAndTrollsRespawn**
> interface{} DungeonsAndTrollsRespawn(ctx, body)
Respawn the Character bound to the logged user.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**interface{}**](interface{}.md)|  | 

### Return type

[**interface{}**](interface{}.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DungeonsAndTrollsSkill**
> interface{} DungeonsAndTrollsSkill(ctx, body)
Use a skill (provided by an item) by the Character bound to the logged user.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**DungeonsandtrollsSkillUse**](DungeonsandtrollsSkillUse.md)|  | 

### Return type

[**interface{}**](interface{}.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DungeonsAndTrollsYell**
> interface{} DungeonsAndTrollsYell(ctx, body)
The Character bound to the logged user yells a messages (visible for everyone).

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**DungeonsandtrollsMessage**](DungeonsandtrollsMessage.md)|  | 

### Return type

[**interface{}**](interface{}.md)

### Authorization

[ApiKeyAuth](../README.md#ApiKeyAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

