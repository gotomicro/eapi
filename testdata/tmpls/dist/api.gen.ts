import { request } from 'umi';




// CallbackWechat
export async function PostApiCallbacksWechat (options?: { [key: string]: any }) {
  return request('/api/callbacks/wechat', {
    method: 'POST',
    
    ...(options || {}),
  });
}


// OssCallback
export async function PostApiCallbacksOss (req: CommonOssCallbackReq,options?: { [key: string]: any }) {
  return request('/api/callbacks/oss', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// GitHubWebhook
export async function PostApiGithubWebhook (options?: { [key: string]: any }) {
  return request('/api/github/webhook', {
    method: 'POST',
    
    ...(options || {}),
  });
}


// GitHubSetup
export async function GetApiGithubSetup (options?: { [key: string]: any }) {
  return request('/api/github/setup', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// OauthUserInfo
export async function GetApiOauthUser (options?: { [key: string]: any }) {
  return request('/api/oauth/user', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// OauthUserInfo
export async function GetApiOauthAppUser (options?: { [key: string]: any }) {
  return request('/api/oauth/appUser', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// Detail
export async function GetApiCmtDetail (options?: { [key: string]: any }) {
  return request('/api/cmt/detail', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// Page
export async function GetApiHomePage (options?: { [key: string]: any }) {
  return request('/api/home/page', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// Trees
export async function GetApiCmtSpaceTrees (options?: { [key: string]: any }) {
  return request('/api/cmt/space-trees', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// ManagersX
export async function GetApiCmtManagers (options?: { [key: string]: any }) {
  return request('/api/cmt/managers', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// MyEmojis
export async function GetApiArticlesMyEmojis (req: ArticleMyEmojisRequest,options?: { [key: string]: any }) {
  return request('/api/articles/-/myEmojis', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// Permission
export async function GetApiSpacesGuidPermission (options?: { [key: string]: any }) {
  return request('/api/spaces/:guid/permission', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// RefState
export async function GetApiReferralsState (req: CommunityRefStateReq,options?: { [key: string]: any }) {
  return request('/api/referrals/state', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// List
export async function GetApiMyCommunityList (options?: { [key: string]: any }) {
  return request('/api/my/community/list', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// Apply
export async function PostApiMyCommunityApply (req: CommunityApplyRequest,options?: { [key: string]: any }) {
  return request('/api/my/community/apply', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// Update
export async function PutApiMyCommunitiesGuid (req: CommunityUpdateRequest,options?: { [key: string]: any }) {
  return request('/api/my/communities/:guid', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// UpdatePrice
export async function PutApiMyCommunitiesGuidPrice (req: CommunityUpdatePriceRequest,options?: { [key: string]: any }) {
  return request('/api/my/communities/:guid/price', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// UpdateBanner
export async function PutApiMyCommunitiesGuidBanner (req: CommunityUpdateBannerRequest,options?: { [key: string]: any }) {
  return request('/api/my/communities/:guid/banner', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// Delete
export async function DeleteApiMyCommunitiesGuid (options?: { [key: string]: any }) {
  return request('/api/my/communities/:guid', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// MemberList
export async function GetApiMyCommunitiesGuidMembers (req: CommunityMemberListRequest,options?: { [key: string]: any }) {
  return request('/api/my/communities/:guid/members', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// UploadBannerToken
export async function PostApiMyCommunitiesGuidUploadBannerToken (req: CommunityUploadTokenRequest,options?: { [key: string]: any }) {
  return request('/api/my/communities/:guid/uploadBannerToken', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// UploadToken
export async function PostApiMyCommunityUploadPicToken (req: CommunityUploadTokenRequest,options?: { [key: string]: any }) {
  return request('/api/my/community/uploadPicToken', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// ApplyCreateCmtAudit
export async function PostApiMyCommunityApplyCreateAudit (options?: { [key: string]: any }) {
  return request('/api/my/community/applyCreateAudit', {
    method: 'POST',
    
    ...(options || {}),
  });
}


// DeleteUser
export async function DeleteApiMyCommunityUsersUid (options?: { [key: string]: any }) {
  return request('/api/my/community/users/:uid', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// UpdateUserStat
export async function PutApiMyCommunityUsersUid (options?: { [key: string]: any }) {
  return request('/api/my/community/users/:uid', {
    method: 'PUT',
    
    ...(options || {}),
  });
}


// UserJoinedCommunities
export async function GetApiMyJoinedCommunities (options?: { [key: string]: any }) {
  return request('/api/my/joined-communities', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// JoinCommunityWithRef
export async function PostApiMyJoinedCommunities (req: CommunityJoinCommunityWithRefReq,options?: { [key: string]: any }) {
  return request('/api/my/joined-communities', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// UserQuitCommunity
export async function DeleteApiMyJoinedCommunitiesGuid (options?: { [key: string]: any }) {
  return request('/api/my/joined-communities/:guid', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// FollowingCreate
export async function PostApiMyFollowingUid (req: MyFollowingCreateReq,options?: { [key: string]: any }) {
  return request('/api/my/following/:uid', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// FollowingDelete
export async function DeleteApiMyFollowingUid (options?: { [key: string]: any }) {
  return request('/api/my/following/:uid', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// BlockList
export async function GetApiMyBlocks (req: MyBlockListReq,options?: { [key: string]: any }) {
  return request('/api/my/blocks', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// BlockCreate
export async function PostApiMyBlocksUid (req: MyBlockCreateReq,options?: { [key: string]: any }) {
  return request('/api/my/blocks/:uid', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// BlockDelete
export async function DeleteApiMyBlocksUid (options?: { [key: string]: any }) {
  return request('/api/my/blocks/:uid', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// NotificationCreate
export async function PostApiMyNotifications (req: MyNotificationCreateReq,options?: { [key: string]: any }) {
  return request('/api/my/notifications', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// NotificationList
export async function GetApiMyNotifications (req: MyNotificationListReq,options?: { [key: string]: any }) {
  return request('/api/my/notifications', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// NotificationAuditList
export async function GetApiMyNotificationsAudits (req: MyNotificationAuditListRequest,options?: { [key: string]: any }) {
  return request('/api/my/notifications/-/audits', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// NotificationAuditPass
export async function PutApiMyNotificationsAuditsAuditIdPass (req: MyNotificationAuditPassRequest,options?: { [key: string]: any }) {
  return request('/api/my/notifications/-/audits/:auditId/pass', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// NotificationAuditReject
export async function PutApiMyNotificationsAuditsAuditIdReject (req: MyNotificationAuditRejectRequest,options?: { [key: string]: any }) {
  return request('/api/my/notifications/-/audits/:auditId/reject', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// NotificationUpdateAll
export async function PutApiMyNotifications (req: MyNotificationUpdateAllReq,options?: { [key: string]: any }) {
  return request('/api/my/notifications', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// NotificationUpdate
export async function PutApiMyNotificationsId (options?: { [key: string]: any }) {
  return request('/api/my/notifications/:id', {
    method: 'PUT',
    
    ...(options || {}),
  });
}


// NotificationUnViewCount
export async function GetApiMyNotificationsUnviewed (options?: { [key: string]: any }) {
  return request('/api/my/notifications/-/unviewed', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// My
export async function GetApiMy (options?: { [key: string]: any }) {
  return request('/api/my', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// UpdateNickname
export async function PutApiMyNickname (req: UserUpdateNicknameRequest,options?: { [key: string]: any }) {
  return request('/api/my/nickname', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// UpdateAvatar
export async function PutApiMyAvatar (req: UserUpdateAvatarRequest,options?: { [key: string]: any }) {
  return request('/api/my/avatar', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// UploadToken
export async function PostApiMyAvatarToken (req: UserUploadTokenRequest,options?: { [key: string]: any }) {
  return request('/api/my/avatarToken', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// UpdatePhone
export async function PutApiMyPhone (req: UserUpdatePhoneRequest,options?: { [key: string]: any }) {
  return request('/api/my/phone', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// UpdateEmail
export async function PutApiMyEmail (req: UserUpdateEmailRequest,options?: { [key: string]: any }) {
  return request('/api/my/email', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// ManageCommunity
export async function PostApiMyManageCommunity (req: UserManageCommunityReq,options?: { [key: string]: any }) {
  return request('/api/my/manageCommunity', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// Referrals
export async function GetApiMyReferrals (req: MyReferralsReq,options?: { [key: string]: any }) {
  return request('/api/my/referrals', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// ReferralsSend
export async function PostApiMyReferralsSend (req: MyReferralsSendReq,options?: { [key: string]: any }) {
  return request('/api/my/referrals/send', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// UserTotal
export async function GetApiUsersNameTotal (options?: { [key: string]: any }) {
  return request('/api/users/:name/total', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// FollowersList
export async function GetApiUsersNameFollowers (req: ProfileFollowersListReq,options?: { [key: string]: any }) {
  return request('/api/users/:name/followers', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// FollowingList
export async function GetApiUsersNameFollowing (req: ProfileFollowingListReq,options?: { [key: string]: any }) {
  return request('/api/users/:name/following', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// UserMedalList
export async function GetApiUsersNameMedals (req: MedalUserMedalListRequest,options?: { [key: string]: any }) {
  return request('/api/users/:name/medals', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// ArticlesList
export async function GetApiUsersNameArticles (req: ProfileArticlesListReq,options?: { [key: string]: any }) {
  return request('/api/users/:name/articles', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// QAList
export async function GetApiUsersNameQuestions (req: ProfileQAListReq,options?: { [key: string]: any }) {
  return request('/api/users/:name/questions', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// Info
export async function GetApiUsersName (options?: { [key: string]: any }) {
  return request('/api/users/:name', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// ListLogos
export async function GetApiCommunitiesRecommendLogos (options?: { [key: string]: any }) {
  return request('/api/communities/recommendLogos', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// ListCovers
export async function GetApiCommunitiesRecommendCovers (options?: { [key: string]: any }) {
  return request('/api/communities/recommendCovers', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// Managers
export async function GetApiCommunitiesGuidManagers (options?: { [key: string]: any }) {
  return request('/api/communities/:guid/managers', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// ListPublic
export async function GetApiCommunities (options?: { [key: string]: any }) {
  return request('/api/communities', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// TemplateList
export async function GetApiCommunitiesTemplates (options?: { [key: string]: any }) {
  return request('/api/communities/-/templates', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// BuyMember
export async function PostApiCommunitiesGuidBuyMember (req: CommunityBuyMemberRequest,options?: { [key: string]: any }) {
  return request('/api/communities/:guid/buy-member', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// UserInfo
export async function GetApiCommunitiesGuidUserInfo (options?: { [key: string]: any }) {
  return request('/api/communities/:guid/user-info', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// TreeChangeSort
export async function PutApiCmtSpaceTreesChangeSort (req: SpaceTreeChangeSortRequest,options?: { [key: string]: any }) {
  return request('/api/cmt/space-trees/change-sort', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// TreeChangeGroup
export async function PutApiCmtSpaceTreesChangeGroup (req: SpaceTreeChangeGroupRequest,options?: { [key: string]: any }) {
  return request('/api/cmt/space-trees/change-group', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// RecommendList
export async function GetApiCmtRecommendList (options?: { [key: string]: any }) {
  return request('/api/cmt/recommendList', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// CollectionGroupCreate
export async function PostApiMyCollectionGroups (req: MyCollectionGroupCreateReq,options?: { [key: string]: any }) {
  return request('/api/my/collection-groups', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// CollectionGroupList
export async function GetApiMyCollectionGroups (options?: { [key: string]: any }) {
  return request('/api/my/collection-groups', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// CollectionGroupDelete
export async function DeleteApiMyCollectionGroupsCgid (options?: { [key: string]: any }) {
  return request('/api/my/collection-groups/:cgid', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// CollectionGroupUpdate
export async function PutApiMyCollectionGroupsCgid (req: MyCollectionGroupUpdateReq,options?: { [key: string]: any }) {
  return request('/api/my/collection-groups/:cgid', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// CollectionList
export async function GetApiMyCollectionGroupsCgidCollections (req: MyCollectionListReq,options?: { [key: string]: any }) {
  return request('/api/my/collection-groups/:cgid/collections', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// CollectionCreate
export async function PostApiMyCollectionGroupsCollections (req: MyCollectionCreateReq,options?: { [key: string]: any }) {
  return request('/api/my/collection-groups/-/collections', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// CollectionDelete
export async function DeleteApiMyCollectionGroupsCollections (req: MyCollectionDeleteReq,options?: { [key: string]: any }) {
  return request('/api/my/collection-groups/-/collections', {
    method: 'DELETE',
    body: req ,
    ...(options || {}),
  });
}


// SearchMember
export async function GetApiCommunitiesSearchMembers (req: CommunitySearchMemberReq,options?: { [key: string]: any }) {
  return request('/api/communities/-/searchMembers', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// Upgrade
export async function PostApiCommunitiesUpgradeEdition (req: EditionUpgradeReq,options?: { [key: string]: any }) {
  return request('/api/communities/-/upgrade-edition', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// CreateGroup
export async function PostApiSpacesGroups (req: SpaceCreateOrUpdateGroupRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/-/groups', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// GroupInfo
export async function GetApiSpacesGroupsGuid (options?: { [key: string]: any }) {
  return request('/api/spaces/-/groups/:guid', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// UpdateGroup
export async function PutApiSpacesGroupsGuid (req: SpaceCreateOrUpdateGroupRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/-/groups/:guid', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// ChangeGroupSort
export async function PutApiSpacesGroupsGuidSort (req: SpaceChangeGroupSortRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/-/groups/:guid/sort', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// DeleteGroup
export async function DeleteApiSpacesGroupsGuid (options?: { [key: string]: any }) {
  return request('/api/spaces/-/groups/:guid', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// GroupMemberList
export async function GetApiSpacesGroupsGuidMembers (req: SpaceGroupMemberListRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/-/groups/:guid/members', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// SearchGroupMember
export async function GetApiSpacesGroupsGuidSearchMembers (req: SpaceSearchMemberRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/-/groups/:guid/searchMembers', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// CreateGroupMember
export async function PostApiSpacesGroupsGuidMembers (req: SpaceCreateMemberRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/-/groups/:guid/members', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// DeleteGroupMember
export async function DeleteApiSpacesGroupsGuidMembers (req: SpaceDeleteMemberRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/-/groups/:guid/members', {
    method: 'DELETE',
    body: req ,
    ...(options || {}),
  });
}


// ApplyGroupMember
export async function PostApiSpacesGroupsGuidApply (req: SpaceApplyGroupMemberRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/-/groups/:guid/apply', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// GroupPermission
export async function GetApiSpacesGroupsGuidPermission (options?: { [key: string]: any }) {
  return request('/api/spaces/-/groups/:guid/permission', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// Create
export async function PostApiSpaces (req: SpaceCreateRequest,options?: { [key: string]: any }) {
  return request('/api/spaces', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// Info
export async function GetApiSpacesGuid (options?: { [key: string]: any }) {
  return request('/api/spaces/:guid', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// Update
export async function PutApiSpacesGuid (req: SpaceUpdateRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/:guid', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// UpdateAttrName
export async function PutApiSpacesGuidUpdateAttrName (req: SpaceUpdateAttrRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/:guid/updateAttrName', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// UpdateAttrIcon
export async function PutApiSpacesGuidUpdateAttrIcon (req: SpaceUpdateAttrRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/:guid/updateAttrIcon', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// UpdateAttrSpaceGroup
export async function PutApiSpacesGuidUpdateAttrGroup (req: SpaceUpdateAttrRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/:guid/updateAttrGroup', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// ChangeSort
export async function PutApiSpacesGuidSort (req: SpaceChangeSortRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/:guid/sort', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// Delete
export async function DeleteApiSpacesGuid (options?: { [key: string]: any }) {
  return request('/api/spaces/:guid', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// MemberList
export async function GetApiSpacesGuidMembers (req: SpaceGroupMemberListRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/:guid/members', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// SearchMember
export async function GetApiSpacesGuidSearchMembers (req: SpaceSearchMemberRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/:guid/searchMembers', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// CreateMember
export async function PostApiSpacesGuidMembers (req: SpaceCreateMemberRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/:guid/members', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// DeleteMember
export async function DeleteApiSpacesGuidMembers (req: SpaceDeleteMemberRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/:guid/members', {
    method: 'DELETE',
    body: req ,
    ...(options || {}),
  });
}


// Emojis
export async function GetApiSpacesGuidEmojis (options?: { [key: string]: any }) {
  return request('/api/spaces/:guid/emojis', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// UploadToken
export async function PostApiSpacesGuidUploadPicToken (req: SpaceUploadTokenRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/:guid/uploadPicToken', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// ImUpload
export async function PostApiSpacesGuidImUpload (req: SpaceImUploadRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/:guid/imUpload', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// ApplyMember
export async function PostApiSpacesGuidApply (req: SpaceApplyMemberRequest,options?: { [key: string]: any }) {
  return request('/api/spaces/:guid/apply', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// ListArticle
export async function GetApiArticles (req: ArticleListArticleRequest,options?: { [key: string]: any }) {
  return request('/api/articles', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// Recommends
export async function GetApiArticlesRecommends (req: ArticleListArticleRequest,options?: { [key: string]: any }) {
  return request('/api/articles/-/recommends', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// SpaceTops
export async function GetApiArticlesSpaceTops (req: ArticleSpaceTopsRequest,options?: { [key: string]: any }) {
  return request('/api/articles/-/spaceTops', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// ListPermission
export async function GetApiArticlesPermissions (req: ArticleListPermissionRequest,options?: { [key: string]: any }) {
  return request('/api/articles/-/permissions', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// ListCovers
export async function GetApiArticlesRecommendCovers (options?: { [key: string]: any }) {
  return request('/api/articles/-/recommendCovers', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// CreateArticle
export async function PostApiArticles (req: ArticleCreateArticleRequest,options?: { [key: string]: any }) {
  return request('/api/articles', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// GetArticle
export async function GetApiArticlesGuid (options?: { [key: string]: any }) {
  return request('/api/articles/:guid', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// GetArticleContent
export async function GetApiArticlesGuidContent (options?: { [key: string]: any }) {
  return request('/api/articles/:guid/content', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// Permission
export async function GetApiArticlesGuidPermission (options?: { [key: string]: any }) {
  return request('/api/articles/:guid/permission', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// GetArticleContentByCreator
export async function GetApiArticlesGuidContentByCreator (options?: { [key: string]: any }) {
  return request('/api/articles/:guid/contentByCreator', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// UpdateArticle
export async function PutApiArticlesGuid (req: ArticleUpdateArticleRequest,options?: { [key: string]: any }) {
  return request('/api/articles/:guid', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// DeleteArticle
export async function DeleteApiArticlesGuid (options?: { [key: string]: any }) {
  return request('/api/articles/:guid', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// SpaceTop
export async function PutApiArticlesGuidSpaceTop (options?: { [key: string]: any }) {
  return request('/api/articles/:guid/spaceTop', {
    method: 'PUT',
    
    ...(options || {}),
  });
}


// CancelSpaceTop
export async function PutApiArticlesGuidCancelSpaceTop (options?: { [key: string]: any }) {
  return request('/api/articles/:guid/cancelSpaceTop', {
    method: 'PUT',
    
    ...(options || {}),
  });
}


// Recommend
export async function PutApiArticlesGuidRecommend (options?: { [key: string]: any }) {
  return request('/api/articles/:guid/recommend', {
    method: 'PUT',
    
    ...(options || {}),
  });
}


// CancelRecommend
export async function PutApiArticlesGuidCancelRecommend (options?: { [key: string]: any }) {
  return request('/api/articles/:guid/cancelRecommend', {
    method: 'PUT',
    
    ...(options || {}),
  });
}


// OpenComment
export async function PutApiArticlesGuidOpenComment (options?: { [key: string]: any }) {
  return request('/api/articles/:guid/openComment', {
    method: 'PUT',
    
    ...(options || {}),
  });
}


// CloseComment
export async function PutApiArticlesGuidCloseComment (options?: { [key: string]: any }) {
  return request('/api/articles/:guid/closeComment', {
    method: 'PUT',
    
    ...(options || {}),
  });
}


// IncreaseEmoji
export async function PutApiArticlesGuidIncreaseEmoji (req: ArticleIncreaseEmojiRequest,options?: { [key: string]: any }) {
  return request('/api/articles/:guid/increaseEmoji', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// DecreaseEmoji
export async function PutApiArticlesGuidDecreaseEmoji (req: ArticleDecreaseEmojiRequest,options?: { [key: string]: any }) {
  return request('/api/articles/:guid/decreaseEmoji', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// CreateComment
export async function PostApiArticlesComments (req: ArticleCreateCommentRequest,options?: { [key: string]: any }) {
  return request('/api/articles/-/comments', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// DeleteComment
export async function DeleteApiArticlesGuidCommentsCommentGuid (options?: { [key: string]: any }) {
  return request('/api/articles/:guid/comments/:commentGuid', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// CommentList
export async function GetApiArticlesGuidComments (req: ArticleTopicCommentListRequest,options?: { [key: string]: any }) {
  return request('/api/articles/:guid/comments', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// ChildCommentList
export async function GetApiArticlesGuidCommentsCommentGuidChildComment (req: ArticleTopicChildCommentListRequest,options?: { [key: string]: any }) {
  return request('/api/articles/:guid/comments/:commentGuid/childComment', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// Create
export async function PostApiQuestions (req: QuestionCreateRequest,options?: { [key: string]: any }) {
  return request('/api/questions', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// List
export async function GetApiQuestions (req: QuestionListRequest,options?: { [key: string]: any }) {
  return request('/api/questions', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// MyEmojis
export async function GetApiQuestionsMyEmojis (req: QuestionMyEmojisRequest,options?: { [key: string]: any }) {
  return request('/api/questions/-/myEmojis', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// Info
export async function GetApiQuestionsGuid (options?: { [key: string]: any }) {
  return request('/api/questions/:guid', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// Update
export async function PutApiQuestionsGuid (req: QuestionUpdateArticleRequest,options?: { [key: string]: any }) {
  return request('/api/questions/:guid', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// Delete
export async function DeleteApiQuestionsGuid (options?: { [key: string]: any }) {
  return request('/api/questions/:guid', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// CreateAnswer
export async function PostApiQuestionsGuidAnswers (req: QuestionCreateRequest,options?: { [key: string]: any }) {
  return request('/api/questions/:guid/answers', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// AnswerList
export async function GetApiQuestionsGuidAnswers (req: QuestionListRequest,options?: { [key: string]: any }) {
  return request('/api/questions/:guid/answers', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// Update
export async function PutApiQuestionsAnswersAnswerGuid (req: QuestionUpdateArticleRequest,options?: { [key: string]: any }) {
  return request('/api/questions/-/answers/:answerGuid', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// IncreaseEmoji
export async function PutApiQuestionsAnswersAnswerGuidIncreaseEmoji (req: QuestionIncreaseEmojiRequest,options?: { [key: string]: any }) {
  return request('/api/questions/-/answers/:answerGuid/increaseEmoji', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// DecreaseEmoji
export async function PutApiQuestionsAnswersAnswerGuidDecreaseEmoji (req: QuestionDecreaseEmojiRequest,options?: { [key: string]: any }) {
  return request('/api/questions/-/answers/:answerGuid/decreaseEmoji', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// AnswerInfo
export async function GetApiQuestionsAnswersAnswerGuid (options?: { [key: string]: any }) {
  return request('/api/questions/-/answers/:answerGuid', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// Delete
export async function DeleteApiQuestionsAnswersAnswerGuid (options?: { [key: string]: any }) {
  return request('/api/questions/-/answers/:answerGuid', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// CreateComment
export async function PostApiQuestionsAnswersAnswerGuidComments (req: QuestionCreateCommentRequest,options?: { [key: string]: any }) {
  return request('/api/questions/-/answers/:answerGuid/comments', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// CommentList
export async function GetApiQuestionsAnswersAnswerGuidComments (req: QuestionCommentListRequest,options?: { [key: string]: any }) {
  return request('/api/questions/-/answers/:answerGuid/comments', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// ChildCommentList
export async function GetApiQuestionsAnswersAnswerGuidCommentsCommentGuidChildComment (req: QuestionChildCommentListRequest,options?: { [key: string]: any }) {
  return request('/api/questions/-/answers/:answerGuid/comments/:commentGuid/childComment', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// DeleteComment
export async function DeleteApiQuestionsAnswersAnswerGuidCommentsCommentGuid (options?: { [key: string]: any }) {
  return request('/api/questions/-/answers/:answerGuid/comments/:commentGuid', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// LikeQuestion
export async function PostApiQuestionsGuidLike (options?: { [key: string]: any }) {
  return request('/api/questions/:guid/like', {
    method: 'POST',
    
    ...(options || {}),
  });
}


// UndoLikeQuestion
export async function DeleteApiQuestionsGuidLike (options?: { [key: string]: any }) {
  return request('/api/questions/:guid/like', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// LikeAnswer
export async function PostApiQuestionsAnswersAnswerGuidLike (options?: { [key: string]: any }) {
  return request('/api/questions/-/answers/:answerGuid/like', {
    method: 'POST',
    
    ...(options || {}),
  });
}


// UndoLikeAnswer
export async function DeleteApiQuestionsAnswersAnswerGuidLike (options?: { [key: string]: any }) {
  return request('/api/questions/-/answers/:answerGuid/like', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// LikeComment
export async function PostApiQuestionsAnswersAnswerGuidCommentsCommentGuidLike (options?: { [key: string]: any }) {
  return request('/api/questions/-/answers/:answerGuid/comments/:commentGuid/like', {
    method: 'POST',
    
    ...(options || {}),
  });
}


// UndoLikeComment
export async function DeleteApiQuestionsAnswersAnswerGuidCommentsCommentGuidLike (options?: { [key: string]: any }) {
  return request('/api/questions/-/answers/:answerGuid/comments/:commentGuid/like', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// Create
export async function PostApiActivities (req: ActivityCreateReq,options?: { [key: string]: any }) {
  return request('/api/activities', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// ListCovers
export async function GetApiActivitiesRecommendCovers (options?: { [key: string]: any }) {
  return request('/api/activities/-/recommendCovers', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// Update
export async function PutApiActivitiesGuid (req: ActivityUpdateReq,options?: { [key: string]: any }) {
  return request('/api/activities/:guid', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// Delete
export async function DeleteApiActivitiesGuid (options?: { [key: string]: any }) {
  return request('/api/activities/:guid', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// List
export async function GetApiActivities (req: ActivityListReq,options?: { [key: string]: any }) {
  return request('/api/activities', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// Info
export async function GetApiActivitiesGuid (options?: { [key: string]: any }) {
  return request('/api/activities/:guid', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// ListJoinedUsers
export async function GetApiActivitiesGuidUsers (req: ActivityListJoinedUsersReq,options?: { [key: string]: any }) {
  return request('/api/activities/:guid/users', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// ListJoined
export async function GetApiJoinedActivities (options?: { [key: string]: any }) {
  return request('/api/joined-activities', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// Join
export async function PostApiJoinedActivitiesGuid (req: ActivityJoinReq,options?: { [key: string]: any }) {
  return request('/api/joined-activities/:guid', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// Quit
export async function DeleteApiJoinedActivitiesGuid (req: ActivityQuitReq,options?: { [key: string]: any }) {
  return request('/api/joined-activities/:guid', {
    method: 'DELETE',
    body: req ,
    ...(options || {}),
  });
}


// Query
export async function GetApiSearch (req: SearchQueryReq,options?: { [key: string]: any }) {
  return request('/api/search', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// Create
export async function PostApiMedals (req: MedalCreateRequest,options?: { [key: string]: any }) {
  return request('/api/medals', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// Update
export async function PutApiMedalsMedalId (req: MedalUpdateRequest,options?: { [key: string]: any }) {
  return request('/api/medals/:medalId', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// Delete
export async function DeleteApiMedalsMedalId (options?: { [key: string]: any }) {
  return request('/api/medals/:medalId', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// List
export async function GetApiMedals (req: MedalMedalListRequest,options?: { [key: string]: any }) {
  return request('/api/medals', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// Award
export async function PostApiMedalsMedalIdMembers (req: MedalAwardRequest,options?: { [key: string]: any }) {
  return request('/api/medals/:medalId/members', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// MemberList
export async function GetApiMedalsMedalIdMembers (req: MedalMemberListRequest,options?: { [key: string]: any }) {
  return request('/api/medals/:medalId/members', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// DeleteMember
export async function DeleteApiMedalsUsersMedalMemberId (options?: { [key: string]: any }) {
  return request('/api/medals/-/users/:medalMemberId', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// UpdateMember
export async function PutApiMedalsUsersMedalMemberId (req: MedalUpdateMemberRequest,options?: { [key: string]: any }) {
  return request('/api/medals/-/users/:medalMemberId', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// UploadToken
export async function PostApiMedalsUploadPicToken (req: MedalUploadTokenRequest,options?: { [key: string]: any }) {
  return request('/api/medals/-/uploadPicToken', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// GoodCreate
export async function PostApiGoods (req: ShopGoodCreateReq,options?: { [key: string]: any }) {
  return request('/api/goods', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// GoodList
export async function GetApiGoods (req: ShopGoodListReq,options?: { [key: string]: any }) {
  return request('/api/goods', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// GoodInfo
export async function GetApiGoodsId (options?: { [key: string]: any }) {
  return request('/api/goods/:id', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// GoodDelete
export async function DeleteApiGoodsId (options?: { [key: string]: any }) {
  return request('/api/goods/:id', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// GoodUpdate
export async function PutApiGoodsId (req: ShopGoodUpdateReq,options?: { [key: string]: any }) {
  return request('/api/goods/:id', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// OrderCreate
export async function PostApiMyOrders (req: ShopOrderCreateReq,options?: { [key: string]: any }) {
  return request('/api/my/orders', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// OrderList
export async function GetApiMyOrders (req: ShopOrderListReq,options?: { [key: string]: any }) {
  return request('/api/my/orders', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// OrderInfo
export async function GetApiMyOrdersSn (options?: { [key: string]: any }) {
  return request('/api/my/orders/:sn', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// OrderDelete
export async function DeleteApiMyOrdersSn (options?: { [key: string]: any }) {
  return request('/api/my/orders/:sn', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// OrderUpdate
export async function PutApiMyOrdersSn (req: ShopOrderUpdateReq,options?: { [key: string]: any }) {
  return request('/api/my/orders/:sn', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// ChargeCreate
export async function PostApiMyCharges (req: ShopChargeCreateReq,options?: { [key: string]: any }) {
  return request('/api/my/charges', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// ChargeDelete
export async function DeleteApiMyChargesSn (options?: { [key: string]: any }) {
  return request('/api/my/charges/:sn', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// ChargeInfo
export async function GetApiMyChargesSn (options?: { [key: string]: any }) {
  return request('/api/my/charges/:sn', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// RefundCreate
export async function PostApiMyRefunds (req: ShopRefundCreateReq,options?: { [key: string]: any }) {
  return request('/api/my/refunds', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// RefundInfo
export async function GetApiMyRefundsSn (req: ShopRefundInfoReq,options?: { [key: string]: any }) {
  return request('/api/my/refunds/:sn', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// CmtUserInfo
export async function GetApiMyCmtUser (options?: { [key: string]: any }) {
  return request('/api/my/cmt/user', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// UpdateUserInfo
export async function PutApiMyCmtUser (req: CommunityUpdateUserInfoRequest,options?: { [key: string]: any }) {
  return request('/api/my/cmt/user', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// List
export async function GetApiEditions (options?: { [key: string]: any }) {
  return request('/api/editions', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// CreateNode
export async function PostApiDriveNodes (req: DriveCreateNodeReq,options?: { [key: string]: any }) {
  return request('/api/drive/nodes', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// ListNode
export async function GetApiDriveNodes (req: DriveListNodeReq,options?: { [key: string]: any }) {
  return request('/api/drive/nodes', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// DeleteNode
export async function DeleteApiDriveNodesNguid (req: DriveDeleteNodeReq,options?: { [key: string]: any }) {
  return request('/api/drive/nodes/:nguid', {
    method: 'DELETE',
    body: req ,
    ...(options || {}),
  });
}


// UpdateNode
export async function PutApiDriveNodesNguid (req: DriveUpdateNodeReq,options?: { [key: string]: any }) {
  return request('/api/drive/nodes/:nguid', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// ListDescendantNode
export async function GetApiDriveAllDescendantNodes (req: DriveListDescendantNodeReq,options?: { [key: string]: any }) {
  return request('/api/drive/all-descendant-nodes', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// ManagerMemberList
export async function GetApiCmtAdminPmsManagersMembers (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/managers/members', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// CreateManagerMember
export async function PostApiCmtAdminPmsManagersMembers (req: PmsCreateManagerMemberReq,options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/managers/members', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// DeleteManagerMember
export async function DeleteApiCmtAdminPmsManagersMembersUid (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/managers/members/:uid', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// RoleList
export async function GetApiCmtAdminPmsRoles (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// UserRoleIds
export async function GetApiCmtAdminPmsRolesUsersUidRoleIds (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles/-/users/:uid/roleIds', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// CreateRole
export async function PostApiCmtAdminPmsRoles (req: PmsCreateRoleRequest,options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// UpdateRole
export async function PutApiCmtAdminPmsRolesRoleId (req: PmsCreateRoleRequest,options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles/:roleId', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// DeleteRole
export async function DeleteApiCmtAdminPmsRolesRoleId (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles/:roleId', {
    method: 'DELETE',
    
    ...(options || {}),
  });
}


// RoleMemberList
export async function GetApiCmtAdminPmsRolesRoleIdMembers (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles/:roleId/members', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// RolePermission
export async function GetApiCmtAdminPmsRolesRoleIdPermissions (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles/:roleId/permissions', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// RoleSpaceGroupPermission
export async function GetApiCmtAdminPmsRolesRoleIdSpaceGroupPermissions (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles/:roleId/spaceGroupPermissions', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// RoleSpacePermission
export async function GetApiCmtAdminPmsRolesRoleIdSpacePermissions (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles/:roleId/spacePermissions', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// RoleSpaceInitPermission
export async function GetApiCmtAdminPmsRolesRoleIdSpacesGuidInitPermissions (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles/:roleId/spaces/:guid/initPermissions', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// RoleSpaceGroupInitPermission
export async function GetApiCmtAdminPmsRolesRoleIdSpaceGroupsGuidInitPermissions (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles/:roleId/spaceGroups/:guid/initPermissions', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// PutRolePermission
export async function PutApiCmtAdminPmsRolesRoleIdPermissions (req: PmsPutRolePermissionRequest,options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles/:roleId/permissions', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// PutRoleSpaceGroupPermission
export async function PutApiCmtAdminPmsRolesRoleIdSpaceGroupPermissions (req: PmsPutRoleSpaceGroupPermissionRequest,options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles/:roleId/spaceGroupPermissions', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// PutRoleSpacePermission
export async function PutApiCmtAdminPmsRolesRoleIdSpacePermissions (req: PmsPutRoleSpacePermissionRequest,options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles/:roleId/spacePermissions', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// CreateRoleMember
export async function PostApiCmtAdminPmsRolesRoleIdMembers (req: PmsCreateRoleMemberRequest,options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles/:roleId/members', {
    method: 'POST',
    body: req ,
    ...(options || {}),
  });
}


// DeleteRoleMember
export async function DeleteApiCmtAdminPmsRolesRoleIdMembers (req: PmsDeleteRoleMemberRequest,options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/pms/roles/:roleId/members', {
    method: 'DELETE',
    body: req ,
    ...(options || {}),
  });
}


// CommunityTotalList
export async function GetApiCmtAdminTotalCommunityKey (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/total/community/:key', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// SpaceTotalList
export async function GetApiCmtAdminTotalSpaceSpaceGuidKey (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/total/space/:spaceGuid/:key', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// ListPage
export async function GetApiCmtAdminLoggerListPage (req: LoggerListPageRequest,options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/logger/listPage', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// EventAndGroupList
export async function GetApiCmtAdminLoggerEventAndGroupList (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/logger/eventAndGroupList', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// Equity
export async function GetApiCmtAdminBillingEquityData (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/billing/equityData', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// ListPage
export async function GetApiCmtAdminCmtGuidBillingListPage (req: BillingListPageRequest,options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/:cmtGuid/billing/listPage', {
    method: 'GET',
    body: req ,
    ...(options || {}),
  });
}


// PlatformAllList
export async function GetApiCmtAdminApps (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/apps', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// Put
export async function PutApiCmtAdminAppsAppId (req: AppPutRequest,options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/apps/:appId', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// Get
export async function GetApiCmtAdminHome (options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/home', {
    method: 'GET',
    
    ...(options || {}),
  });
}


// Put
export async function PutApiCmtAdminHome (req: HomePutReq,options?: { [key: string]: any }) {
  return request('/api/cmtAdmin/home', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// Put
export async function PutApiIntgsGithubAppId (req: AppPutRequest,options?: { [key: string]: any }) {
  return request('/api/intgs/github/appId', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}


// QueryFigures
export async function PutApiIntgsFigures (req: ThirdQueryFiguresReq,options?: { [key: string]: any }) {
  return request('/api/intgs/figures', {
    method: 'PUT',
    body: req ,
    ...(options || {}),
  });
}




type MedalMemberListRequest = {
    
    currentPage: number; // 当前页数
    
}

type HomePutReq = {
    
    defaultPageByNotLogin: string; // 未登录用户默认访问页面
    
    isSetHome: boolean; // 是否启用首页
    
    isSetBanner: boolean; // 是否启用banner
    
    articleSortByLogin: number; // 登录用户推荐内容排序规则
    
    articleHotShowSum: number; // 展示热门帖子的数量
    
    articleHotShowWithLatest: number; // 展示近期多少天内创建的帖子
    
    bannerLink: string; // banner的跳转链接
    
    articleSortByNotLogin: number; // 未登录用户推荐内容排序规则
    
    activityLatestShowSum: number; // 展示近期的活动数量
    
    bannerImg: string; // 启用首页banner
    
    bannerTitle: string; // banner文案
    
    defaultPageByNewUser: string; // 新用户注册默认访问页面
    
    defaultPageByLogin: string; // 登录用户默认访问页面
    
}

type ActivityCreateReq = {
    
    content: string; // 
    
    address: string; // 
    
    cover: string; // 
    
    name: string; // 
    
    guestUids: number[]; // 
    
    location: string; // 
    
    fee: number; // 
    
    isOnline: number; // 
    
    isFree: number; // 
    
    spaceGuid: string; // 
    
    description: string; // 
    
    startTime: number; // 
    
    endTime: number; // 
    
    signUpStartTime: number; // 
    
    signUpEndTime: number; // 
    
}

type ActivityJoinReq = {
    
    uid: number; // 
    
}

type Orderv1OrderGood = {
    
    payTotal: number; // 一般是 payPrice * count - 优惠。目前没有优惠信息
    
    specList: string; // 规格
    
    payPrice: number; // 商品实际支付价格(拼团商品适用)
    
    goodId: number; // 商品主表ID
    
    goodSkuId: number; // 商品ID
    
    goodFreightFee: number; // 商品的运费
    
    title: string; // 商品名称
    
    goodFreightWay: number; // 商品运费方式
    
    groupPrice: number; // 拼团价格 just价格，一般和实际支付价格一致
    
    buyerId: number; // buyerId
    
    priceType: number; // 价格类型
    
    id: number; // 主键
    
    captainPrice: number; // 团长价格 just价格，一般和实际支付价格一致
    
    num: number; // 购买数量
    
    cmtGuid: string; // 社区Guid
    
    price: number; // 商品价格 just价格
    
    goodType: number; // 商品类型
    
    orderId: number; // 订单ID
    
    img: string; // 图片
    
}

type BillingListPageRequest = {
    
    currentPage: number; // 当前页数
    
}

type MyBlockCreateReq = {
    
    pagination: Commonv1Pagination; // 
    
}

type SpaceGroupMemberListRequest = {
    
    currentPage: number; // 当前页数
    
}

type ProfileFollowingListReq = {
    
    sort: string; // 排序字符串
    
    currentPage: number; // 当前页数
    
    pageSize: number; // 每页总数
    
}

type SpaceSearchMemberRequest = {
    
    keyword: string; // 当前页数
    
}

type SpaceUpdateRequest = {
    
    visibilityLevel: number; // visibilityLevel
    
    isAllowReadMemberList: boolean; // 
    
    spaceOptions: Spacev1SpaceOption[]; // 
    
    name: string; // SpaceGroupGuid        string                 `json:"spaceGroupGuid" binding:"required" label:"组"`
    
    iconType: number; // 
    
    icon: string; // 
    
    spaceType: number; // 空间类型
    
    spaceLayout: number; // 空间布局
    
}

type ArticleCreateArticleRequest = {
    
    name: string; // 
    
    spaceGuid: string; // 
    
    parentGuid: string; // 
    
    content: any; // 
    
    headImage: string; // 
    
    imageUrls: string[]; // 
    
}

type QuestionUpdateArticleRequest = {
    
    content: string; // 
    
    imageUrls: string[]; // 
    
    name: string; // 
    
}

type ActivityListJoinedUsersReq = {
    
    currentPage: number; // 当前页数
    
}

type CommunityMemberListRequest = {
    
    currentPage: number; // 当前页数
    
}

type Commonv1Pagination = {
    
    currentPage: number; // 当前页数
    
    pageSize: number; // 每页总数
    
    sort: string; // 排序字符串
    
    total: number; // 总数
    
}

type PmsCreateRoleRequest = {
    
    name: string; // 
    
}

type PmsPutRoleSpacePermissionRequest = {
    
    list: Pmsv1SpacePmsItem[]; // 
    
}

type MedalMedalListRequest = {
    
    currentPage: number; // 当前页数
    
}

type ShopOrderCreateReq = {
    
    buyerEmail: string; // 买家邮箱
    
    buyerAvatar: string; // 冗余字段，买家头像
    
    totalAmount: number; // 订单总价格
    
    prePayAmount: number; // 预付款
    
    orderGoodList: Orderv1OrderGood[]; // 商品列表
    
    buyerName: string; // 买家姓名
    
    buyerPhone: string; // 买家电话
    
    remark: string; // 备注
    
    goodAmount: number; // 商品合计
    
    ext: Orderv1OrderExtend; // 扩展字段
    
}

type ActivityUpdateReq = {
    
    location: string; // 
    
    content: string; // 
    
    address: string; // 
    
    name: string; // 
    
    fee: number; // 
    
    guestUids: number[]; // 
    
    cover: string; // 
    
    isOnline: number; // 
    
    isFree: number; // 
    
    endTime: number; // 
    
    signUpStartTime: number; // 
    
    signUpEndTime: number; // 
    
    description: string; // 
    
    startTime: number; // 
    
}

type CommunityUploadTokenRequest = {
    
    fileName: string; // 
    
    size: number; // 
    
}

type MyNotificationAuditRejectRequest = {
    
    opReason: string; // 管理员理由
    
}

type MedalUpdateRequest = {
    
    name: string; // 
    
    icon: string; // 
    
    desc: string; // 
    
}

type MedalAwardRequest = {
    
    uids: number[]; // 
    
    startTime: number; // 
    
    endTime: number; // 
    
    type: number; // 
    
}

type PmsDeleteRoleMemberRequest = {
    
    uids: number[]; // 
    
}

type ArticleListArticleRequest = {
    
    spaceGuid: string; // 
    
    currentPage: number; // 当前页数
    
    sort: number; // 排序值
    
}

type ArticleCreateCommentRequest = {
    
    guid: string; // 
    
    commentGuid: string; // 
    
    content: string; // 
    
}

type ShopGoodCreateReq = {
    
    galleryList: string[]; // 
    
    cidList: number[]; // 
    
    artistDesc: string; // 
    
    worksDesc: string; // 
    
    name: string; // 
    
    saleTime: number; // 
    
    payType: number; // 
    
    price: number; // 
    
    html: string; // 
    
    wechatHtml: string; // 
    
    cmtGuid: string; // 
    
    endTime: number; // 
    
    freightId: number; // 
    
    imageSpecId: number; // 
    
    originPrice: number; // 
    
    preMarkdown: string; // 
    
    markdown: string; // 
    
    title: string; // 
    
    subTitle: string; // 
    
    showPrice: number; // 
    
    skuList: Goodv1GoodSku[]; // 
    
    preHtml: string; // 
    
    cover: string; // 
    
    showOriginPrice: number; // 
    
    onlineTime: number; // 
    
    stock: number; // 
    
    groupSaleNum: number; // 
    
    freightFee: number; // 
    
    baseSaleNum: number; // 
    
    homeCover: string; // 
    
    saleNum: number; // 
    
    specList: Goodv1GoodSpecInfo[]; // 
    
}

type ShopOrderUpdateReq = {
    
    state: number; // 订单状态
    
}

type CommunityUpdateUserInfoRequest = {
    
    nickname: string; // 
    
    position: string; // 
    
}

type MyFollowingCreateReq = {
    
    currentPage: number; // 当前页数
    
    pageSize: number; // 每页总数
    
    sort: string; // 排序字符串
    
}

type SpaceApplyMemberRequest = {
    
    reason: string; // 
    
}

type ShopChargeCreateReq = {
    
    chargeMethod: number; // 
    
    amount: number; // 
    
    subject: string; // 
    
    describe: string; // 
    
    orderSn: string; // 
    
}

type PmsCreateManagerMemberReq = {
    
    uids: number[]; // 
    
}

type AppPutRequest = {
    
    enable: boolean; // 
    
}

type CommunityApplyRequest = {
    
    visibilityLevel: number; // 
    
    templateId: number; // 
    
    name: string; // 
    
}

type EditionUpgradeReq = {
    
    duration: number; // 需要Increase或Decrease的Duration，如果无需续期，则此字段可以为零值，暂时只允许Increase
    
    durationUnit: number; // 时长单位,天、月、年，如无需续期，则此字段可以为零值
    
    editionId: number; // 需要升级的版本ID，如果无需升级Edition，则可以此字段可以为零值
    
    chargeMethod: number; // 支付方式，不能为空
    
}

type SpaceCreateMemberRequest = {
    
    addUids: number[]; // 
    
}

type Spacev1SpaceOption = {
    
    name: string; // 名称
    
    spaceOptionId: number; // 可选项
    
    value: number; // 可选项值
    
    spaceOptionType: number; // 可选项类型
    
}

type QuestionChildCommentListRequest = {
    
    currentPage: number; // 
    
}

type PmsCreateRoleMemberRequest = {
    
    uids: number[]; // 
    
}

type CommunityUpdateRequest = {
    
    name: string; // 
    
    logo: string; // 
    
    visibilityLevel: number; // 
    
}

type CommunityBuyMemberRequest = {
    
    chargeMethod: number; // 支付方式，不能为空
    
    memberType: number; // 年度会员 为 1， 月度会员 为 2
    
}

type MedalUploadTokenRequest = {
    
    fileName: string; // 
    
    size: number; // 
    
}

type DriveListDescendantNodeReq = {
    
    currentPage: number; // 当前页数
    
    pageSize: number; // 每页总数
    
    sort: string; // 排序字符串
    
    guid: string; // 自己guid
    
    spaceGuid: string; // 空间guid
    
}

type PmsPutRoleSpaceGroupPermissionRequest = {
    
    list: Pmsv1SpaceGroupPmsItem[]; // 
    
}

type LoggerListPageRequest = {
    
    currentPage: number; // 当前页数
    
    event: number; // 
    
    group: number; // 
    
    operateUid: number; // 
    
    targetUid: number; // 
    
}

type CommunityUpdateBannerRequest = {
    
    img: string; // 
    
    title: string; // 
    
}

type MyReferralsSendChannel = {
    
    channel: string; // 发送通道，email、sms
    
    ref: string; // 邀请码
    
    receiver: string; // 邮箱或电话
    
}

type CommunityRefStateReq = {
    
    ref: string; // 邀请码
    
}

type CommunityJoinCommunityWithRefReq = {
    
    ref: string; // 邀请码
    
}

type ShopGoodUpdateReq = {
    
}

type ArticleIncreaseEmojiRequest = {
    
    emojiId: number; // 
    
}

type QuestionCreateCommentRequest = {
    
    commentGuid: string; // 如果为空，说明是一级评论
    
    content: string; // 
    
}

type MedalUpdateMemberRequest = {
    
    type: number; // 
    
    startTime: number; // 
    
    endTime: number; // 
    
}

type ArticleMyEmojisRequest = {
    
    guids: string[]; // 
    
}

type ArticleListPermissionRequest = {
    
    guids: string[]; // 
    
}

type SearchQueryReq = {
    
    keyword: string; // Keyword 进行查询字符串
    
    bizType: number; // BizType 业务类型，如果为0表示搜索全部业务类型
    
    currentPage: number; // 当前页数
    
    pageSize: number; // 每页总数
    
    sort: string; // 排序字符串
    
}

type Goodv1GoodSku = {
    
    updatedUid: number; // 
    
    price: number; // 
    
    stock: number; // 
    
    weight: number; // 
    
    saleNum: number; // 
    
    specValueSign: string; // 规格值标识，用于买东西的时候，校验是否选择了所
    
    contractId: number; // 
    
    code: string; // 
    
    specSign: string; // 规格标识
    
    whiteLimitNum: number; // 
    
    createdUid: number; // 
    
    originPrice: number; // 
    
    cover: string; // 
    
    groupSaleNum: number; // 
    
    note: string; // 
    
    id: number; // 
    
    goodId: number; // 
    
    specList: Goodv1GoodSkuSpecInfo[]; // 
    
    title: string; // 
    
    status: number; // 
    
    cmtGuid: string; // 
    
    issueType: number; // 
    
}

type MyNotificationUpdateAllReq = {
    
    status: number; // 
    
}

type CommunitySearchMemberReq = {
    
    keyword: string; // 当前页数
    
    spaceGuid: string; // 空间guid，可以为空
    
    currentPage: number; // 当前页数
    
    pageSize: number; // 每页总数
    
    sort: string; // 排序字符串
    
}

type UserManageCommunityReq = {
    
    currentPage: number; // 当前页数
    
}

type QuestionDecreaseEmojiRequest = {
    
    emojiId: number; // 
    
}

type ActivityListReq = {
    
    currentPage: number; // 当前页数
    
    creatorUid: string; // 活动创建人，为空表示不限定创建人
    
}

type UserUpdateAvatarRequest = {
    
    url: string; // 
    
}

type UserUploadTokenRequest = {
    
    fileName: string; // 
    
    size: number; // 
    
}

type QuestionCreateRequest = {
    
    spaceGuid: string; // 
    
    content: string; // 
    
    imageUrls: string[]; // 
    
    name: string; // 
    
}

type SpaceImUploadRequest = {
    
    fileName: string; // 
    
    type: number; // 
    
}

type ArticleTopicChildCommentListRequest = {
    
    currentPage: number; // 
    
}

type CommunityUpdatePriceRequest = {
    
    isSetPrice: boolean; // 
    
    annualPrice: number; // 
    
    annualOriginPrice: number; // 
    
}

type Pmsv1SpaceGroupPmsItem = {
    
    guid: string; // guid 信息
    
    name: string; // 名称
    
    list: Commonv1PmsItem[]; // space group 里面的权限列表
    
}

type Goodv1GoodSpecValueInfo = {
    
    cmtGuid: string; // 
    
    Id: number; // 
    
    Name: string; // 
    
}

type DriveCreateNodeReqItem = {
    
    name: string; // 
    
    size: number; // 
    
}

type UserUpdateNicknameRequest = {
    
    nickname: string; // 
    
}

type SpaceApplyGroupMemberRequest = {
    
    reason: string; // 
    
}

type DriveListNodeReq = {
    
    sort: string; // 排序字符串
    
    currentPage: number; // 当前页数
    
    parentGuid: string; // 不能和SpaceGuid同事为空
    
    spaceGuid: string; // 空间guid
    
    pageSize: number; // 每页总数
    
}

type MyCollectionGroupUpdateReq = {
    
    title: string; // 
    
    desc: string; // 
    
}

type DriveCreateNodeReq = {
    
    spaceGuid: string; // 不能为空
    
    type: number; // 不能为空
    
    nodes: DriveCreateNodeReqItem[]; // 不能为空
    
    parentGuid: string; // 不能和SpaceGuid同时为空
    
}

type ArticleTopicCommentListRequest = {
    
    currentPage: number; // 
    
}

type UserUpdateEmailRequest = {
    
    email: string; // 
    
}

type MyCollectionListReq = {
    
    currentPage: number; // 当前页数
    
    pageSize: number; // 每页总数
    
    sort: string; // 排序字符串
    
}

type MyNotificationCreateReq = {
    
    targetId: string; // 对象ID
    
    uids: number[]; // 用户ID列表
    
    link: string; // 消息外链
    
    ctime: number; // 附属数据
//Meta []byte `protobuf:"bytes,7,opt,name=meta,proto3" json:"meta,omitempty"`
//可选. 通知时间. 如果是0值则自动取当前时间
    
    type: number; // 消息类型
    
    targetType: number; // 对象类型
    
}

type ShopRefundCreateReq = {
    
    orderSn: string; // 
    
    chargeSn: string; // 
    
    amount: number; // 
    
    reason: string; // 
    
}

type PmsPutRolePermissionRequest = {
    
    list: Commonv1PmsItem[]; // 
    
}

type ThirdQueryFiguresReq = {
    
    key: string; // 唯一Key名, 比如 "member_growth"
    
    st: number; // 开始时间戳, 秒
    
    et: number; // 终止时间戳, 秒
    
    type: string; // 集成类型
    
    name: string; // 显示名称, 比如 "用户增长"
    
}

type SpaceUploadTokenRequest = {
    
    fileName: string; // 
    
    type: number; // 
    
    size: number; // 
    
}

type QuestionMyEmojisRequest = {
    
    guids: string[]; // 
    
}

type SpaceCreateRequest = {
    
    visibilityLevel: number; // visibilityLevel
    
    spaceGroupGuid: string; // 
    
    name: string; // 
    
    iconType: number; // 
    
    icon: string; // 
    
    spaceType: number; // 空间类型
    
    spaceThirdType: number; // 空间第三方类型
    
    spaceLayout: number; // 空间布局
    
}

type Pmsv1SpacePmsItem = {
    
    name: string; // 名称
    
    list: Commonv1PmsItem[]; // space里面的权限列表
    
    guid: string; // guid 信息
    
}

type MyNotificationListReq = {
    
    sort: string; // 排序字符串
    
    currentPage: number; // 当前页数
    
    pageSize: number; // 每页总数
    
}

type SpaceChangeGroupSortRequest = {
    
    spaceGroupGuid: string; // 
    
    afterSpaceGroupGuid: string; // 
    
}

type SpaceUpdateAttrRequest = {
    
    spaceGroupGuid: string; // 
    
    name: string; // 
    
    icon: string; // 
    
}

type Goodv1GoodSpecInfo = {
    
    cmtGuid: string; // 
    
    Id: number; // 
    
    Name: string; // 
    
    ValueList: Goodv1GoodSpecValueInfo[]; // 
    
}

type DriveDeleteNodeReq = {
    
    spaceGuid: string; // 
    
}

type UserUpdatePhoneRequest = {
    
    phone: string; // 
    
}

type MyCollectionGroupCreateReq = {
    
    title: string; // 
    
    desc: string; // 
    
}

type ProfileQAListReq = {
    
    sort: string; // 排序字符串
    
    currentPage: number; // 当前页数
    
    pageSize: number; // 每页总数
    
}

type Orderv1OrderExtend = {
    
    orderId: number; // 订单ID
    
    receiverName: string; // 收货人姓名
    
    receiverPhone: string; // 收获人手机
    
    receiverAddrRegion: string; // 收货人地址地区
    
    receiverAddrDetail: string; // 收货人地址地区
    
    trackingNo: string; // 物流单号
    
    expressId: number; // 物流公司id，默认为0 代表不需要物流
    
    expressName: string; // 物流公司
    
}

type Commonv1PmsItem = {
    
    flag: number; // 符号，现在是switch，也可能后续是select
    
    needUpgrade: number; // 是否需要升级
    
    upgradeDesc: string; // 升级描述
    
    actionName: string; // 标识
    
    title: string; // 标题
    
    desc: string; // 描述
    
}

type MyNotificationAuditListRequest = {
    
    sort: string; // 排序字符串
    
    currentPage: number; // 当前页数
    
    pageSize: number; // 每页总数
    
}

type MyReferralsReq = {
    
}

type SpaceDeleteMemberRequest = {
    
    delUids: number[]; // 
    
}

type ShopGoodListReq = {
    
    currentPage: number; // 当前页数
    
    pageSize: number; // 每页总数
    
    sort: string; // 排序字符串
    
}

type ProfileArticlesListReq = {
    
    currentPage: number; // 当前页数
    
    pageSize: number; // 每页总数
    
    sort: string; // 排序字符串
    
}

type MyCollectionDeleteReq = {
    
    collectionGroupIds: number[]; // 
    
    guid: string; // 
    
    type: number; // 
    
}

type MyCollectionCreateReq = {
    
    collectionGroupIds: number[]; // 
    
    guid: string; // 
    
    type: number; // 
    
}

type ShopRefundInfoReq = {
    
    chargeSn: string; // 
    
    refundSn: string; // 
    
}

type CommonOssCallbackReq = {
    
    spaceGuid: string; // 
    
    fileGuid: string; // 
    
    uid: number; // 
    
    fileType: number; // 
    
    bucket: string; // 
    
    object: string; // 
    
    size: number; // 
    
    cmtGuid: string; // 
    
}

type ProfileFollowersListReq = {
    
    sort: string; // 排序字符串
    
    currentPage: number; // 当前页数
    
    pageSize: number; // 每页总数
    
}

type ArticleSpaceTopsRequest = {
    
    spaceGuid: string; // 
    
}

type SpaceTreeChangeSortRequest = {
    
    type: number; // 
    
    afterSpaceGuid: string; // 
    
    spaceGuid: string; // 
    
    afterSpaceGroupGuid: string; // 
    
    spaceGroupGuid: string; // 
    
}

type SpaceCreateOrUpdateGroupRequest = {
    
    visibilityLevel: number; // 
    
    isOpenReadMemberList: boolean; // 如果打开，属于这个分组下的用户，可以看到用户列表
    
    name: string; // 
    
    iconType: number; // 
    
    icon: string; // 
    
}

type QuestionIncreaseEmojiRequest = {
    
    emojiId: number; // 
    
}

type QuestionCommentListRequest = {
    
    currentPage: number; // 
    
}

type ActivityQuitReq = {
    
    uid: number; // 
    
}

type ShopOrderListReq = {
    
    currentPage: number; // 当前页数
    
    pageSize: number; // 每页总数
    
    sort: string; // 排序字符串
    
    start: number; // 
    
    limit: number; // 
    
}

type ArticleDecreaseEmojiRequest = {
    
    emojiId: number; // 
    
}

type QuestionListRequest = {
    
    spaceGuid: string; // 
    
    currentPage: number; // 当前页数
    
    sort: number; // 排序值
    
}

type Goodv1GoodSkuSpecInfo = {
    
    id: number; // 
    
    name: string; // 
    
    valueId: number; // 
    
    valueName: string; // 
    
    valueImg: string; // 
    
    cmtGuid: string; // 
    
}

type SpaceChangeSortRequest = {
    
    afterSpaceGuid: string; // 
    
    spaceGuid: string; // 
    
}

type MedalCreateRequest = {
    
    name: string; // 
    
    icon: string; // 
    
    desc: string; // 
    
}

type SpaceTreeChangeGroupRequest = {
    
    spaceGuid: string; // 
    
    afterSpaceGroupGuid: string; // 
    
}

type ArticleUpdateArticleRequest = {
    
    imageUrls: string[]; // 
    
    headImage: string; // 
    
    name: string; // 
    
    content: string; // 
    
    contentHtml: string; // 
    
}

type MyNotificationAuditPassRequest = {
    
    opReason: string; // 管理员理由
    
}

type MedalUserMedalListRequest = {
    
    currentPage: number; // 当前页数
    
}

type MyBlockListReq = {
    
    pagination: Commonv1Pagination; // 
    
}

type DriveUpdateNodeReq = {
    
    spaceGuid: string; // 
    
    name: string; // 新文件名, 如无需修改可不赋值
    
    parentGuid: string; // 新父guid, 如无需修改可不赋值
    
}
