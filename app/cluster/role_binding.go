package cluster

import (
	"context"
	"errors"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Cluster) CreateClusterRoleBinding(ctx context.Context, username string) error {
	name := fmt.Sprintf("cluster-role-binding-%s", username)

	_, err := c.clientset.RbacV1().ClusterRoleBindings().Create(ctx, &rbacv1.ClusterRoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "User",
				Name:     username,
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "ro-user-role",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}, v1.CreateOptions{})

	if err != nil {
		var a *k8serrors.StatusError

		if !errors.As(err, &a) {
			return fmt.Errorf("create crb: %v", err)
		} else {
			switch a.ErrStatus.Code {
			case 409:
				return nil
			default:
				return err
			}
		}
	}

	return err
}
