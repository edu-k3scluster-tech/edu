package cluster

import (
	"context"
	"errors"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Cluster) CreateClusterRoleBinding(ctx context.Context, username, role string) error {
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
			Name:     role,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}, v1.CreateOptions{})

	return handleRoleErr(err)
}

func (c *Cluster) CreateRoleBinding(ctx context.Context, username, namespace, role string) error {
	name := fmt.Sprintf("role-binding-%s", username)

	_, err := c.clientset.RbacV1().RoleBindings(namespace).Create(ctx, &rbacv1.RoleBinding{
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
			Name:     role,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}, v1.CreateOptions{})

	return handleRoleErr(err)
}

func handleRoleErr(err error) error {
	if err == nil {
		return nil
	}
	var a *k8serrors.StatusError

	if !errors.As(err, &a) {
		return fmt.Errorf("create rb: %v", err)
	} else {
		switch a.ErrStatus.Code {
		case 409:
			return nil
		default:
			return err
		}
	}
}
