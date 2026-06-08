package apperrors

// Estos mapas traducen nombres de constraints de PostgreSQL a errores de
// dominio. Los nombres deben coincidir exactamente con los definidos en las
// migraciones (internal/db/migrations/). Si un constraint se renombra en una
// migración futura, el mapeo deja de funcionar silenciosamente: el error
// llegará al handler como un 500 genérico en lugar del 404/409 esperado.

var UserConstraints = map[string]error{
	"users_email_key": ErrDuplicateEmail,
}

var CategoriesConstraints = map[string]error{
	"categories_name_key": ErrDuplicateName,
}

var CulturalWorksConstraints = map[string]error{
	"cultural_works_title_key":        ErrDuplicateName,
	"cultural_works_category_id_fkey": ErrCategoryNotFound,
}

var FocusTypesConstraints = map[string]error{
	"focus_types_name_key": ErrDuplicateName,
}

var GroupsConstraints = map[string]error{
	"groups_created_by_fkey": ErrUserNotFound,
	"groups_work_id_fkey":    ErrCulturalWorkNotFound,
}

var GroupsFocusTypesConstraints = map[string]error{
	"groups_focus_types_focus_type_id_fkey": ErrFocusTypeNotFound,
	"groups_focus_types_group_id_fkey":      ErrGroupNotFound,
}

var UserInterestsConstraints = map[string]error{
	"user_interests_category_id_fkey": ErrCategoryNotFound,
	"user_interests_profile_id_fkey":  ErrUserNotFound,
}

var ProfilesConstraints = map[string]error{
	"user_profiles_user_id_key":  ErrProfileDuplicateUser,
	"user_profiles_user_id_fkey": ErrUserNotFound,
}

var UsersFocusTypesConstraints = map[string]error{
	"users_focus_types_focus_type_id_fkey": ErrFocusTypeNotFound,
	"users_focus_types_profile_id_fkey":    ErrUserNotFound,
}
