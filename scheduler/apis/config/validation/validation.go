//
//
package validation

import (
	"errors"
	"fmt"
	"github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"github.com/turtacn/cloud-prophet/scheduler/helper/sets"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidatePolicy checks for errors in the Config
// It does not return early so that it can find as many errors as possible
func ValidatePolicy(policy config.Policy) error {
	var validationErrors []error

	priorities := make(map[string]config.PriorityPolicy, len(policy.Priorities))
	for _, priority := range policy.Priorities {
		if priority.Weight <= 0 || priority.Weight >= config.MaxWeight {
			validationErrors = append(validationErrors, fmt.Errorf("Priority %s should have a positive weight applied to it or it has overflown", priority.Name))
		}
		validationErrors = append(validationErrors, validateCustomPriorities(priorities, priority))
	}

	if extenderErrs := validateExtenders(field.NewPath("extenders"), policy.Extenders); len(extenderErrs) > 0 {
		validationErrors = append(validationErrors, extenderErrs.ToAggregate().Errors()...)
	}

	if policy.HardPodAffinitySymmetricWeight < 0 || policy.HardPodAffinitySymmetricWeight > 100 {
		validationErrors = append(validationErrors, field.Invalid(field.NewPath("hardPodAffinitySymmetricWeight"), policy.HardPodAffinitySymmetricWeight, "not in valid range [0-100]"))
	}
	return utilerrors.NewAggregate(validationErrors)
}

// validateExtenders validates the configured extenders for the Scheduler
func validateExtenders(fldPath *field.Path, extenders []config.Extender) field.ErrorList {
	allErrs := field.ErrorList{}
	binders := 0
	extenderManagedResources := sets.NewString()
	for i, extender := range extenders {
		path := fldPath.Index(i)
		if len(extender.PrioritizeVerb) > 0 && extender.Weight <= 0 {
			allErrs = append(allErrs, field.Invalid(path.Child("weight"),
				extender.Weight, "must have a positive weight applied to it"))
		}
		if extender.BindVerb != "" {
			binders++
		}
		for j, resource := range extender.ManagedResources {
			managedResourcesPath := path.Child("managedResources").Index(j)
			errs := validateExtendedResourceName(v1.ResourceName(resource.Name))
			for _, err := range errs {
				allErrs = append(allErrs, field.Invalid(managedResourcesPath.Child("name"),
					resource.Name, fmt.Sprintf("%+v", err)))
			}
			if extenderManagedResources.Has(resource.Name) {
				allErrs = append(allErrs, field.Invalid(managedResourcesPath.Child("name"),
					resource.Name, "duplicate extender managed resource name"))
			}
			extenderManagedResources.Insert(resource.Name)
		}
	}
	if binders > 1 {
		allErrs = append(allErrs, field.Invalid(fldPath, fmt.Sprintf("found %d extenders implementing bind", binders), "only one extender can implement bind"))
	}
	return allErrs
}

// validateCustomPriorities validates that:
// 1. RequestedToCapacityRatioRedeclared custom priority cannot be declared multiple times,
// 2. LabelPreference/ServiceAntiAffinity custom priorities can be declared multiple times,
// however the weights for each custom priority type should be the same.
func validateCustomPriorities(priorities map[string]config.PriorityPolicy, priority config.PriorityPolicy) error {
	verifyRedeclaration := func(priorityType string) error {
		if existing, alreadyDeclared := priorities[priorityType]; alreadyDeclared {
			return fmt.Errorf("Priority %q redeclares custom priority %q, from:%q", priority.Name, priorityType, existing.Name)
		}
		priorities[priorityType] = priority
		return nil
	}
	verifyDifferentWeights := func(priorityType string) error {
		if existing, alreadyDeclared := priorities[priorityType]; alreadyDeclared {
			if existing.Weight != priority.Weight {
				return fmt.Errorf("%s  priority %q has a different weight with %q", priorityType, priority.Name, existing.Name)
			}
		}
		priorities[priorityType] = priority
		return nil
	}
	if priority.Argument != nil {
		if priority.Argument.LabelPreference != nil {
			if err := verifyDifferentWeights("LabelPreference"); err != nil {
				return err
			}
		} else if priority.Argument.RequestedToCapacityRatioArguments != nil {
			if err := verifyRedeclaration("RequestedToCapacityRatio"); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("No priority arguments set for priority %s", priority.Name)
		}
	}
	return nil
}

// validateExtendedResourceName checks whether the specified name is a valid
// extended resource name.
func validateExtendedResourceName(name v1.ResourceName) []error {
	var validationErrors []error
	for _, msg := range validation.IsQualifiedName(string(name)) {
		validationErrors = append(validationErrors, errors.New(msg))
	}
	if len(validationErrors) != 0 {
		return validationErrors
	}
	return validationErrors
}
