package cluster

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
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
			Name:     "ro-cluster-role",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}, v1.CreateOptions{})
	return err
}
