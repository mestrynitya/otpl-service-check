from setuptools import setup, find_packages

def parse_requirements(filename):
    """ load requirements from a pip requirements file """
    lineiter = (line.strip() for line in open(filename))
    return [line for line in lineiter if line and not line.startswith("#")]

setup(
    name="otpl-service-check",
    description="A selection of Nagios plugins to monitor services hosted in OpenTable Mesos.",
    long_description=open('README.rst').read(),
    version="1.1.1", # NB: Version is duplicated in otpl-service-check.
    packages=find_packages(),
    author='OpenTable Architecture Team',
    author_email='archteam@opentable.onmicrosoft.com',
    url="https://github.com/opentable/otpl-service-check",
    scripts=["otpl-service-check"],
    license="Apache 2",
    install_requires=parse_requirements("requirements.txt"),
    include_package_data=True,
    classifiers=[
      'Development Status :: 4 - Beta',
      'Environment :: Console',
      'Intended Audience :: System Administrators',
      'License :: OSI Approved :: Apache Software License',
      'Topic :: System :: Monitoring',
      'Topic :: System :: Networking :: Monitoring'
    ]
)
